# Lambda Function URL

This project provides a way of emulating a Lambda Function URL outside of AWS.

## Getting started

### Running the Lambda

Start your Lambda in your favourite execution engine (e.g., Kubernetes).
Make sure to set the `AWS_LAMBDA_RUNTIME_API` and `_LAMBDA_SERVER_ROOT` environment variables.

### Running the gateway

Run this application somewhere close (e.g. as a Kubernetes sidecar), and set the following environment variables:

* `APP_UPSTREAM` - TCP address of the Lambda (e.g. `localhost:1234`)
* `APP_TIMEOUT` - Number of seconds after which the request should be cancelled (default: `30`).
