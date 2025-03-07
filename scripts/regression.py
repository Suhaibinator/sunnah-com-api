import os
import requests
import json
from deepdiff import DeepDiff
import pymysql  # or your preferred database library
from dotenv import load_dotenv

# Load the environment variables from the .env file

load_dotenv(".env.prod_connect")
resp = requests.get("http://127.0.0.1:5000")
print(resp)
# Database connection setup
def get_db_connection():
    connection = pymysql.connect(
        host="{MYSQL_HOST}".format(**os.environ),
        user="{MYSQL_USER}".format(**os.environ),
        password="{MYSQL_PASSWORD}".format(**os.environ),
        db="{MYSQL_DATABASE}".format(**os.environ),
        cursorclass=pymysql.cursors.DictCursor,
    )
    return connection


# Function to fetch data from the database
def fetch_all(query):
    connection = get_db_connection()
    try:
        with connection.cursor() as cursor:
            cursor.execute(query)
            result = cursor.fetchall()
    finally:
        connection.close()
    return result


# Function to compare API responses
def compare_responses(url, params=None, headers=None, paginate=False):
    original_api_base_url = "http://localhost:5000"
    new_api_base_url = "http://localhost:8084"

    def fetch_paginated_data(base_url):
        page = 1
        all_data = []
        while True:
            params_with_pagination = params.copy() if params else {}
            params_with_pagination.update({"page": page, "limit": 100})
            response = requests.get(f"{base_url}{url}", params=params_with_pagination, headers=headers)
            if response.status_code != 200:
                print(f"Error {response.status_code} for URL {url} on page {page}")
                break
            data = response.json()
            if not data:
                break
            all_data.extend(data)
            page += 1
        return all_data

    if paginate:
        original_json = fetch_paginated_data(original_api_base_url)
        new_json = fetch_paginated_data(new_api_base_url)
    else:
        original_response = requests.get(f"{original_api_base_url}{url}", params=params, headers=headers)
        new_response = requests.get(f"{new_api_base_url}{url}", params=params, headers=headers)

        original_json = original_response.json()
        new_json = new_response.json()

    diff = DeepDiff(original_json, new_json, significant_digits=5)

    if diff:
        print(f"Differences found for URL {url} with parameters {params}:")
        print(json.dumps(diff, indent=4))
    else:
        print(f"No differences found for URL {url} with parameters {params}.")


# Main function to run tests
def run_regression_tests():
    headers = {"Authorization": "Bearer your_token"}  # if authentication is required

    # Test /v1/collections
    compare_responses("/v1/collections", headers=headers, paginate=True)

    # Fetch all collection names
    collections = fetch_all("SELECT name FROM HadithCollection ORDER BY collectionID")
    for collection in collections:
        name = collection["name"]

        # Test /v1/collections/<string:name>
        compare_responses(f"/v1/collections/{name}", headers=headers)

        # Test /v1/collections/<string:name>/books
        compare_responses(f"/v1/collections/{name}/books", headers=headers, paginate=True)

        # Fetch all books for the collection with status=4
        books = fetch_all(f"SELECT ourBookID FROM Book WHERE collection='{name}' AND status=4")
        for book in books:
            bookNumber = book["ourBookID"]

            # Test /v1/collections/<string:name>/books/<string:bookNumber>
            compare_responses(f"/v1/collections/{name}/books/{bookNumber}", headers=headers)

            # Test /v1/collections/<string:collection_name>/books/<string:bookNumber>/hadiths
            compare_responses(f"/v1/collections/{name}/books/{bookNumber}/hadiths", headers=headers, paginate=True)

            # Fetch all chapters for the book
            chapters = fetch_all(f"SELECT babID FROM Chapter WHERE collection='{name}' AND arabicBookID='{bookNumber}'")
            for chapter in chapters:
                chapterId = chapter["babID"]

                # Test /v1/collections/<string:collection_name>/books/<string:bookNumber>/chapters/<float:chapterId>
                compare_responses(f"/v1/collections/{name}/books/{bookNumber}/chapters/{chapterId}", headers=headers)

            # Test /v1/collections/<string:collection_name>/books/<string:bookNumber>/chapters
            compare_responses(f"/v1/collections/{name}/books/{bookNumber}/chapters", headers=headers, paginate=True)

        # Fetch hadith numbers for the collection
        hadiths = fetch_all(f"SELECT hadithNumber FROM Hadith WHERE collection='{name}'")
        for hadith in hadiths:
            hadithNumber = hadith["hadithNumber"]

            # Test /v1/collections/<string:collection_name>/hadiths/<string:hadithNumber>
            compare_responses(f"/v1/collections/{name}/hadiths/{hadithNumber}", headers=headers)

    # Fetch all URNs
    urns = fetch_all("SELECT DISTINCT arabicURN FROM Hadith UNION SELECT DISTINCT englishURN FROM Hadith")
    for urn in urns:
        urn_value = urn.get("arabicURN") or urn.get("englishURN")

        # Test /v1/hadiths/<int:urn>
        compare_responses(f"/v1/hadiths/{urn_value}", headers=headers)

    # Test /v1/hadiths/random
    compare_responses("/v1/hadiths/random", headers=headers)


if __name__ == "__main__":
    run_regression_tests()
