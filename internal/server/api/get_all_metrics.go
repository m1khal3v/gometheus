package api

import (
	"html/template"
	"net/http"
)

const pageTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>gometheus</title>
</head>
<body>
<table>
    <tr>
        <th>Name</th>
        <th>Type</th>
        <th>Value</th>
    </tr>
    {{ range .}}
        <tr>
            <td>{{ .GetName }}</td>
            <td>{{ .GetType }}</td>
            <td>{{ .GetStringValue }}</td>
        </tr>
    {{ end}}
</table>
</body>
</html>
`

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
