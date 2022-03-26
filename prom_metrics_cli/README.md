## Command line interface for extracting and visualising prometheus metrics by querying prometheus http api. 

This project is written mainly in go 1.17 using multithreaded http requests and python for visualising the results with matplotlib.
Matplotlib figures are saved in "./plot/figures" as png files with ascending id.

## Install:
```bash
chmod +x install.sh
./install.sh
go run main.go --help
```

You can find an example script in "gateway_function_invocation_total_range_query.sh". You can run it with:
```bash
./gateway_function_invocation_total_range_query.sh
```

