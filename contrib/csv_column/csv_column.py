#!/usr/bin/env python3
import csv
import json
import sys

def main():
    if len(sys.argv) != 2 and len(sys.argv) != 1:
        print(f"Usage: {sys.argv[0]} [column_name]", file=sys.stderr)
        print("  If column_name is provided, prints values for that column", file=sys.stderr)
        print("  If no column_name is provided, prints all column names", file=sys.stderr)
        sys.exit(1)

    sys.stdin.reconfigure(encoding='utf-8-sig')
    reader = csv.reader(sys.stdin)

    try:
        header = next(reader)
    except StopIteration:
        print("Error: CSV file is empty", file=sys.stderr)
        sys.exit(1)

    if len(sys.argv) == 1:
        for col in header:
            print(json.dumps(col))
        return

    column_name = sys.argv[1]
    try:
        col_index = header.index(column_name)
    except ValueError:
        print(f"Column '{column_name}' not found in CSV header", file=sys.stderr)
        sys.exit(1)

    for row in reader:
        if col_index < len(row):
            print(json.dumps(row[col_index]))

if __name__ == "__main__":
    main()
