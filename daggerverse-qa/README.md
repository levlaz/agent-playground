# Daggerverse QA Agent

A Dagger module for automated Quality Assurance testing of Daggerverse modules.

## Features

- Automated testing of Daggerverse modules
- Validation of module interfaces and functionality
- Test report generation

## Requirements

- Dagger CLI (latest version)

## Usage

Run the QA agent using the Dagger CLI:

```shell
dagger call -m github.com/levlaz/agent-playground/daggerverse-qa do-qa --module=path/to/your/module
```

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
- experiment with dagger install github.com/kpenfound/dag/firecrawl@v1.0.2

## License

MIT License
