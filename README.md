# VA API Go Client

This is an example client for working with [VA APIs](https://developer.va.gov/) written in Go.

To run the application, make sure you have credentials. If you need them, you can [request them on the developer portal](https://developer.va.gov/apply). When filling out a callback URL, make sure the localhost port matches whatever port you use in the run command below. This example sets port 5000, but change it accordingly.

With credentials in hand, set the following environment variables:
```
VA_CLIENT={your client id from the API Platform team}
VA_SECRET={your client secret from the API Platform team}
```

Then you can run the app with the following command:
```
go build && ENV=LOCAL PORT=5000 VA_API='https://dev-api.va.gov' ./va-api-client
```
