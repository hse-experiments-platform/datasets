import sys

import numpy as np
import requests
import typing
import fire
import pandas as pd

allowed_types = ['string', 'int', 'float', 'dropped', 'categorial']


def convert(dataset_id: int, rows_count: int, schema: typing.List[typing.Tuple[str, str]], token: str,
            result_file_path: str, base_url: str):
    dataset_id = int(dataset_id)
    rows_count = int(rows_count)
    token = str(token)
    base_url = str(base_url)
    result_file_path = str(result_file_path)

    table = [[]]
    for t in schema:
        (name, type) = t
        if type not in allowed_types:
            print(f"type not in {allowed_types}, got {type}")
            sys.exit(3)
        table[0].append(name)

    resp = requests.get(f"http://{base_url}/api/v1/datasets/{dataset_id}/rows?limit={rows_count}",
                        headers={
                            "Authorization": token})

    rows = resp.json()['rows']

    for row in rows:
        cols = row['columns']
        if len(cols) != len(table[0]):
            print(f'got invalid row, expected columns {len(table[0])}, got {len(cols)}')
        table = np.append(table, [cols], axis=0)

    df = pd.DataFrame(data=table[1:,0:],columns=table[0,0:])

    i = 0
    for (name, type) in schema:
        try:
            if type == 'int':
                df[name] = df[name].astype(int)
            elif type == 'float':
                df[name] = df[name].astype(float)
            elif type == 'string' or type == 'categorial':
                df[name] = df[name].astype(str)
            elif type == 'dropped':
                df = df.drop(name, axis=1)
        except ValueError as err:
            print(f"cannot convert column {name} to type {type}")
            sys.exit(2)
        i += 1

    df.to_csv(result_file_path, sep=',', index=False)


if __name__ == "__main__":
    fire.Fire(convert)
