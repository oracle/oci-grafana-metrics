import requests
import re

# Function to fetch and process regions
def fetch_and_process_regions(custom_regions=None):
    # Step 1: Fetch the raw content of the file from the URL
    url = "https://raw.githubusercontent.com/oracle/oci-typescript-sdk/refs/heads/master/lib/common/lib/region.ts"
    response = requests.get(url)
    file_content = response.text

    # Step 2: Find all lines that contain 'Region.register'
    matches = re.findall(r'Region\.register\("([^"]+)"', file_content)

    # Step 3: Sort the matches and remove duplicates
    unique_sorted_matches = sorted(set(matches))

    # Step 4: Add custom regions (if provided)
    if custom_regions:
        unique_sorted_matches.extend(custom_regions)
        unique_sorted_matches = sorted(set(unique_sorted_matches))  # Remove duplicates after adding custom regions

    # Step 5: Format the result to mimic the output of the original command
    regions = ', '.join([f"'{match}'" for match in unique_sorted_matches])
    regions = f"export const regions = [{regions}]"

    # Step 6: Write the result to the file 'regionlist.ts'
    with open('./src/regionlist.ts', 'w') as file:
        file.write(regions)

    print("Result successfully written to './src/regionlist.ts'")

fetch_and_process_regions()