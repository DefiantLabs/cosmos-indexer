pipeline {
    agent {
        label 'BUILD'
    }
    triggers {
        githubPush()
    }

    stages {
        stage('Checkout Code') {
            steps {
                script {
                    env.DOCKER_APP = "${JOB_NAME}"
                    env.DOCKER_NET_NAME = "vpcbr"
                    env.POSTGRES_CONTAINER = "${env.DOCKER_APP}_postgres"
                    env.REDIS_CONTAINER = "redis"
                    env.IMAGE_NAME = "${env.DOCKER_APP}:${env.BUILD_NUMBER}"
                }
                checkout scm
            }
        }
//        stage('Build Docker Image') {
//            steps {
//                sh "docker build -t ${env.IMAGE_NAME} ."
//            }
//        }
        stage('Deploy') {
            agent { label "jenkins-agent-2" }
            steps {
                deployApplication()
            }
        }
    }
    post {
        always {
            cleanUp()
        }
    }
}

void deployApplication() {
    createDockerNetwork()
    runPostgres()
    runRedis()
    runMongo()
    runApplication()
}

void createDockerNetwork() {
    def networkStatus = sh(script: "docker network ls | grep ${env.DOCKER_NET_NAME} && echo true || echo false", returnStdout: true).trim()
    if (networkStatus.contains("false")) {
        sh script: "docker network create --driver=bridge --subnet=10.5.0.0/16 --gateway=10.5.0.1 ${env.DOCKER_NET_NAME}",
           label: "Create docker network"
    }
}

void runPostgres() {
    def pgStatus = sh(script: "docker ps -a | grep ${env.POSTGRES_CONTAINER} && echo true || echo false", returnStdout: true).trim()
    if (pgStatus.contains("false")) {
        sh """
            docker run -d --name ${env.POSTGRES_CONTAINER} \
                --restart unless-stopped \
                -v /etc/localtime:/etc/localtime:ro \
                -e POSTGRES_USER=taxuser \
                -e POSTGRES_PASSWORD=password \
                -e POSTGRES_DB=postgres \
                -p 5437:5432 \
                --ip 10.5.0.8 \
                --network ${env.DOCKER_NET_NAME} \
                postgres:15-alpine
        """
        // Wait for DB to be ready
        sh """
            while ! docker exec ${env.POSTGRES_CONTAINER} pg_isready -q -h localhost -p 5432 -U postgres; do
                sleep 1
            done
        """
    }
}

void runRedis() {
    def redisStatus = sh(script: "docker ps -a | grep ${env.REDIS_CONTAINER} && echo true || echo false", returnStdout: true).trim()
    if (redisStatus.contains("false")) {
        sh """
            docker run -d --name ${env.REDIS_CONTAINER} \
                --restart unless-stopped \
                -p 6381:6379 \
                --ip 10.5.0.10 \
                --network ${env.DOCKER_NET_NAME} \
                redis
        """
    }
}

void runMongo() {
    def mongoStatus = sh(script: "docker ps -a | grep mongodb && echo true || echo false", returnStdout: true).trim()
    if (mongoStatus.contains("false")) {
        sh """
            docker run -d --name mongodb \
                --restart unless-stopped \
                --shm-size=1g \
                -e MONGO_INITDB_DATABASE=search_indexer \
                -e MONGO_INITDB_ROOT_USERNAME=admin \
                -e MONGO_INITDB_ROOT_PASSWORD=password \
                -v db:/data/db \
                -p 27017-27019:27017-27019 \
                --ip 10.5.0.21 \
                --network ${env.DOCKER_NET_NAME} \
                mongo:5.0.2
        """
    }
}

void runApplication() {
    def appStatus = sh(script: "docker ps -a | grep ${env.DOCKER_APP} && echo true || echo false", returnStdout: true).trim()
    if (appStatus.contains("true")) {
        sh script: "docker rm -fv ${env.DOCKER_APP}", label: "Remove ${env.DOCKER_APP} container"
    }
    sh "perl -pi -e 's/pg_container/${env.POSTGRES_CONTAINER}/g; s/image/${env.IMAGE_NAME}/g' .env"
    sh "docker-compose -f docker-compose-depl.yaml up -d"
}

void cleanUp() {
    try {
        cleanWs()
        def directories = ["@tmp", "@script", "@script@tmp", "@2", "@2@tmp"]
        directories.each { dirSuffix ->
            dir("${env.WORKSPACE}${dirSuffix}") {
                deleteDir()
            }
        }
    } catch (Exception e) {
        echo 'Error cleaning dirs: ' + e
    }
}
