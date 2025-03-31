# Daggerverse QA

An automated QA tool for Daggerverse modules.

## Features

- Automated testing of Daggerverse modules
- Validation of module interfaces and functionality
- Test report generation

## Usage 

This tool automatically tests Daggerverse modules and generates HTML reports for each module in the `/qa` directory.

Run the QA agent using the Dagger shell:

```shell
. $FIRECRAWL_API_KEY lev@levlaz.org           
                 $SURGE_TOKEN daggerverse-qa | do-qa | export  
                 ./qa
```

## Publishing Reports

The generated reports are automatically published to surge.sh. The process:

1. All module reports are stored in the `/qa` directory
2. An `index.html` is automatically generated with a summary table of all tested modules
3. The entire `/qa` directory is published to surge.sh using the Dagger pipeline


## Development

To contribute to this module:

1. Clone the repository
2. Make your changes
3. Test locally using `dagger call test`
4. Submit a pull request

## TODO 

- Consinstently do more than one thing at a time 
- Do matrix builds with multiple versions of Dagger Engine
- Set up some cron jobs to run this all the time 
- Get notifications going with novu, maybe notify people when api change causes issues?
- multi model 

## License

MIT License
