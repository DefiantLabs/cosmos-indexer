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

EPOCHS_DAY_ALL_QUERY = "SELECT epoch_number, start_height FROM epoches WHERE identifier='day' ORDER BY start_height ASC;"
BLOCKS_BY_HEIGHT_QUERY = "SELECT id, height FROM blocks WHERE height=%s;"
EVENTS_BY_HEIGHT_AND_SOURCE_QUERY = "SELECT COUNT(*) FROM taxable_event WHERE block_id=%s AND source=0;"

# An naive/unoptimized implementation to find counts for epoch reward events
# Does the following:
# 1. Finds all epochs
# 2. Finds the block ID for each epoch height
# 3. Counts the reward for each block id
# 4. Reassociates the block rewards back with its epoch
# 5. Outputs the counts for each epoch
if __name__ == "__main__":
    env = get_env()
    os.makedirs("./output", exist_ok=True)

    conn = psycopg2.connect(f"dbname={env['db_name']} user={env['user']} host={env['host']} password={env['password']} port={env['port']}")
    try:
        epochs = []
        with conn.cursor() as cur:
            cur.execute(EPOCHS_DAY_ALL_QUERY)
            for record in cur.fetchall():
                epochs.append({
                    "epoch_number": record[0],
                    "start_height": record[1]
                })

        json.dump(epochs, open("output/epochs_tracked.json", 'w'), indent=4)
        blocks = []
        heights = [int(epoch["start_height"]) for epoch in epochs]
        with conn.cursor() as cur:
            for height in heights:
                cur.execute(BLOCKS_BY_HEIGHT_QUERY, (height,))
                rec = cur.fetchone()
                if rec is None:
                    print(height)
                    print(BLOCKS_BY_HEIGHT_QUERY % height)
                    continue
                blocks.append({
                    "id": rec[0],
                    "height": rec[1]
                })

        json.dump(blocks, open("output/blocks_tracked.json", 'w'), indent=4)
        events_by_block_height_counts = {}
        for block in blocks:
            with conn.cursor() as cur:
                cur.execute(EVENTS_BY_HEIGHT_AND_SOURCE_QUERY, (block["id"],))
                rec = cur.fetchone()
                events_by_block_height_counts[block["height"]] = rec[0]
        json.dump(events_by_block_height_counts, open("output/events_by_block_id_tracked.json", 'w'), indent=4)

        epoch_number_counts = []
        available_heights = events_by_block_height_counts.keys()

        for epoch in epochs:
            if epoch["start_height"] in available_heights:
                epoch_number_counts.append({
                    "epoch_number": epoch["epoch_number"],
                    "count": events_by_block_height_counts[epoch["start_height"]]
                })
            else:
                epoch_number_counts.append({
                    "epoch_number": epoch["epoch_number"],
                    "count": 0
                })

        json.dump(epoch_number_counts, open("output/epoch_counts.json", 'w'), indent=4)


    except Exception as err:
        print(err)
        traceback.print_exc()
    finally:
        conn.close()
