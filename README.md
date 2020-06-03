[![New Relic Experimental header](https://github.com/newrelic/open-source-office/raw/master/examples/categories/images/Experimental.png)](https://github.com/newrelic/open-source-office/blob/master/examples/categories/index.md#category-new-relic-experimental)

# nri-winservices
![](https://github.com/newrelic/nri-winservices/workflows/PullRequestAndMergeMaster/badge.svg)
![](https://github.com/newrelic/nri-winservices/workflows/Release/badge.svg)



New Relic's Windows Services integration brings data about the services running on your Windows hosts into our platform.
 You can check the state and start mode of each service, find out which hosts are running a service, add services to 
 workloads, set up alerts for services, and more.

Our integration is bundled with the Windows agent: if you are already monitoring Windows hosts on New Relic, you just 
need to enable the integration to get Windows services data into our platform.

To get data from Windows hosts a reduced version of the Prometheus exporter for 
Windows is used. It exposes Prometheus metrics on the port specified in the agent configuration, which are collected by the 
integration, transformed into entities and metrics, filtered, and then set to New Relic.


## Installation

This integration will be included as a bundle in the future releases of the Agent. It will be not enabled by default.

## Building 
In order to build the project it is available a PowerShell script: `win_build.ps1` that will take care of building the 
project and the supported version of the integration placing binaries in the folder `/target/bin`
Note that only Windows is supported.

## Testing

Once built, the integration can be tested running `nri-winservices.exe` placed in the `./target/bin` directory. 
It will spin up automatically the exported with the provided settings. 

## Support

This project is not supported yet and it is currently under development

## Contributing
Contributions to improve nri-winservices are encouraged! Keep in mind when you submit your pull request, you'll need to
 sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.
To execute our corporate CLA, which is required if your contribution is on behalf of a company, or if you have any
 questions, please drop us an email at open-source@newrelic.com.

## License
nri-winservices is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
