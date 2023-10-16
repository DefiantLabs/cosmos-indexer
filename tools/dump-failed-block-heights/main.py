import psycopg2
import os
import json
import traceback

def get_env():
    ret = {
        "host": os.environ.get("DB_HOST", ""),
        "password": os.environ.get("DB_PASSWORD", ""),
        "user": os.environ.get("DB_USER", ""),
        "port": os.environ.get("DB_PORT", ""),
        "db_name": os.environ.get("DB_NAME", "")
    }

    if any([ret[x] == "" for x in ret]):
        raise Exception("Must provide env vars")

    return ret

DUMP_FAILED_BLOCKS_QUERY = "SELECT height FROM failed_blocks ORDER BY height ASC;"

if __name__ == "__main__":
    env = get_env()
    os.makedirs("./output", exist_ok=True)

    conn = psycopg2.connect(f"dbname={env['db_name']} user={env['user']} host={env['host']} password={env['password']} port={env['port']}")
    try:
        heights = []
        with conn.cursor() as cur:
            cur.execute(DUMP_FAILED_BLOCKS_QUERY)
            for record in cur.fetchall():
                heights.append(record[0])

        json.dump(heights, open("output/failed_heights.json", 'w'), indent=4)

    except Exception as err:
        print(err)
        traceback.print_exc()
    finally:
        conn.close()
