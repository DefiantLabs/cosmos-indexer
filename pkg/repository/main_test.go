package repository

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
)

var (
	postgresConn *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dockerPool, err := dockertest.NewPool("")
	if err != nil {
		log.Err(err).Msgf("Could not connect to docker: %s", err)
	}

	resourcePostgres := initializePostgres(ctx, dockerPool, newPostgresConfig())

	postgresManualMigration(ctx)

	//Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	purgeResources(dockerPool, resourcePostgres)

	os.Exit(code)
}

func initializePostgres(ctx context.Context, dockerPool *dockertest.Pool, cfg *postgresConfig) *dockertest.Resource {
	resource, err := dockerPool.Run(cfg.Repository, cfg.Version, cfg.EnvVariables)
	if err != nil {
		log.Err(err).Msgf("Could not start resource: %s", err)
	}

	var dbHostAndPort string

	err = dockerPool.Retry(func() error {
		var dbHost string

		gitlabCIHost := os.Getenv("DATABASE_HOST")

		if gitlabCIHost != "" {
			dbHost = gitlabCIHost
		} else {
			dbHost = "localhost"
		}

		port := resource.GetPort(cfg.PortID)
		dbHostAndPort = fmt.Sprintf("%s:%s", dbHost, port)

		dsn := cfg.getConnectionString(dbHostAndPort)

		postgresConn, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("connect: %v", err)
		}

		if err = postgresConn.Ping(ctx); err != nil {
			return fmt.Errorf("ping: %v", err)
		}

		return nil
	})
	if err != nil {
		log.Err(err).Msgf("Could not connect to database: %s", err)
	}
	log.Info().Msgf(strings.Join(cfg.getFlywayMigrationArgs(dbHostAndPort), " "))
	cmd := exec.Command("/usr/local/bin/flyway", cfg.getFlywayMigrationArgs(dbHostAndPort)...)

	err = cmd.Run()
	if err != nil {
		log.Err(err).Msgf("There are errors in migrations: %v", err)
	}
	return resource
}

type postgresConfig struct {
	Repository   string
	Version      string
	EnvVariables []string
	PortID       string
	DB           string
}

func newPostgresConfig() *postgresConfig {
	return &postgresConfig{
		Repository: "postgres",
		Version:    "14.1-alpine",
		EnvVariables: []string{
			"POSTGRES_PASSWORD=password123",
			"POSTGRES_DB=db",
			"listen_addresses = '*'",
		},
		PortID: "5432/tcp",
		DB:     "db",
	}
}

func (p *postgresConfig) getConnectionString(dbHostAndPort string) string {
	return fmt.Sprintf("postgresql://postgres:password123@%v/%s?sslmode=disable", dbHostAndPort, p.DB)
}

func (p *postgresConfig) getFlywayMigrationArgs(dbHostAndPort string) []string {
	return []string{
		"-user=postgres",
		"-password=password123",
		"-locations=filesystem:../../migrations",
		fmt.Sprintf("-url=jdbc:postgresql://%v/%s", dbHostAndPort, p.DB),
		"migrate",
	}
}

func purgeResources(dockerPool *dockertest.Pool, resources ...*dockertest.Resource) {
	for i := range resources {
		if err := dockerPool.Purge(resources[i]); err != nil {
			log.Err(err).Msgf("Could not purge resource: %s", err)
		}
		err := resources[i].Expire(1)
		if err != nil {
			log.Err(err).Msgf("%s", err)
		}
	}
}

func postgresManualMigration(ctx context.Context) {
	// TODO add this into migration files
	migrations := make([]string, 0)

	queryTxes := `
	create table txes
	(
		id                             bigserial primary key,
		hash                           text,
		code                           bigint,
		block_id                       bigint,
		signatures                     bytea[],
		timestamp                      timestamp with time zone,
		memo                           text,
		timeout_height                 bigint,
		extension_options              text[],
		non_critical_extension_options text[],
		auth_info_id                   bigint,
		tx_response_id                 bigint
	);
	
	create unique index idx_txes_hash
		on txes (hash);`
	migrations = append(migrations, queryTxes)

	queryBlocks := `create table blocks
	(
		id                       bigserial primary key,
		time_stamp               timestamp with time zone,
		height                   bigint,
		chain_id                 bigint,
		proposer_cons_address_id bigint,
		tx_indexed               boolean,
		block_events_indexed     boolean,
		block_hash               text
	)`
	migrations = append(migrations, queryBlocks)

	queryFees := `create table fees
		(
			id               bigserial primary key,
			tx_id            bigint,
			amount           numeric(78),
			denomination_id  bigint,
			payer_address_id bigint
		);`
	migrations = append(migrations, queryFees)

	for _, query := range migrations {
		_, err := postgresConn.Exec(ctx, query)
		if err != nil {
			log.Err(err).Msgf("couldn't manual postgres migration: %s", err.Error())
			return
		}
	}
}
