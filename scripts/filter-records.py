import json
import re

def parse_size(size_str):
    """
    Converts a size string (e.g., '6.697kB', '4.5GB') into bytes for comparison and sorting.
    Supports B, KB, MB, GB, TB.
    """
    units = {
        'B': 1,
        'KB': 1024,
        'MB': 1024**2,
        'GB': 1024**3,
        'TB': 1024**4
    }
    
    match = re.match(r"([\d\.]+)\s*([a-zA-Z]+)", size_str.strip())
    if not match:
        return 0
    
    number, unit = match.groups()
    return float(number) * units.get(unit.upper(), 1)

def filter_and_sort_json(input_filename, output_filename, threshold_gb=4):
    threshold_bytes = threshold_gb * (1024**3)
    
    try:
        with open(input_filename, 'r') as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"Error: {input_filename} not found.")
        return

    filtered_and_sorted = sorted(
        [entry for entry in data if parse_size(entry.get('bucket', '0B')) > threshold_bytes],
        key=lambda x: parse_size(x.get('bucket', '0B'))
    )

    with open(output_filename, 'w') as f:
        json.dump(filtered_and_sorted, f, indent=4)
    
    print(f"Processed {len(filtered_and_sorted)} entries.")

if __name__ == "__main__":
    filter_and_sort_json('input.json', 'output.json')
    