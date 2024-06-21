package api

import (
	"html/template"
	"net/http"
)

func (container Container) GetAllMetrics(writer http.ResponseWriter, request *http.Request) {
	template, err := template.New("get_all_metrics").Parse(pageTemplate)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)
	err = template.Execute(writer, container.manager.GetAll())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

const pageTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <style>
    .metrics_table {
        overflow: auto;
        width: 100%;
    }
    
    .metrics_table table {
        border: 1px dashed #b7b7b7;
        height: 100%;
        width: 60%;
        table-layout: auto;
        border-collapse: collapse;
        border-spacing: 1px;
        text-align: left;
        display: table;
        margin-right: auto;
        margin-left: auto;
    }
    
    .metrics_table th {
        border: 1px dashed #b7b7b7;
        background-color: #eceff1;
        color: #000000;
        padding: 10px;
    }
    
    .metrics_table td {
        border: 1px dashed #b7b7b7;
        padding: 10px;
    }
    
    .metrics_table tr:nth-child(even) td {
        background-color: #ffffff;
        color: #000000;
    }
    
    .metrics_table tr:nth-child(odd) td {
        background-color: #ffffff;
        color: #000000;
    }
    </style>
    <title>gometheus</title>
</head>
<body>
<div class="metrics_table" role="region" tabindex="0">
    <table>
        <thead>
            <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Value</th>
            </tr>
        </thead>
        <tbody>
        {{ range .}}
            <tr>
                <td>{{ .GetName }}</td>
                <td>{{ .GetType }}</td>
                <td>{{ .GetStringValue }}</td>
            </tr>
        {{ end}}
        </tbody>
    </table>
</div>
</body>
</html>
`
