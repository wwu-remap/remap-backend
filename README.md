# ReMAP Backend

A simple backend solution for ReMAP (Remote Monitoring Application in Psychiatry) written in Go.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

* _Go_ must be installed on your machine
* A working MongoDB instance with a fresh database `remap_dev`
* Create a collection `auth` and insert a test user with `subjectId` and `password`:
```
// user for testing
{
    "subjectId" : "Test1234",
    "password" : "secret"
}
```

### Installing

Just build the binary by typing `make` into your console and start it aferwards with your own parameters

The parameters for starting the backend are: `server adress` `mongodb adress` `mongodb database name` `api key`

```
> make
> ./bin/remap-server localhost:8080 mongodb://localhost:27017 remap_dev e7382215-1cb3-4e00-8183-29ed63955d99
```

Sample requests:

```
// GET iOS tasks (there should be some in the database ;) )
curl --location --request GET 'localhost:8080/tasks?ios' \
--header 'x-api-key: e7382215-1cb3-4e00-8183-29ed63955d99' \
--header 'Authorization: Basic VGVzdDEyMzQ6c2VjcmV0'

// POST location event
curl --location --request POST 'localhost:8080/events' \
--header 'x-api-key: e7382215-1cb3-4e00-8183-29ed63955d99' \
--header 'Authorization: Basic VGVzdDEyMzQ6c2VjcmV0' \
--header 'Content-Type: application/json' \
--data-raw '{
    "type": "location",
    "lat": "56.234",
    "lon": "8.32456"
}'

// Upload audio file
curl --location --request POST '192.168.11.8:8080/upload' \
--header 'x-api-key: e7382215-1cb3-4e00-8183-29ed63955d99' \
--header 'Authorization: Basic VGVzdDEyMzQ6c2VjcmV0' \
--header 'Content-Type: audio/x-m4a' \
--data-binary '@/path/to/your/audiofile.m4a'
```

## Built With

* [Maven](https://maven.apache.org/) - Dependency Management

## Contributing

Please contact me if you want to contribute.

## Authors

* **Daniel Emden** - *Initial work* - [danielemden](https://github.com/danielemden)
* **Sven Willner** - *Initial work* - [swillner](https://github.com/swillner)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

* @swillner for helping me writing the POC
