### This program uses google cloud platform (GCP) translation api to convert English into Chinese

Remember to set environment variable of `GOOGLE_APPLICATION_CREDENTIALS` to JSON file contains account key.

Currently set api request each time character limit to 7000, truncate by new line character(`\n`)

### Usage

`translate file_name`

`file_name` is expected to text file contains English to translate.
