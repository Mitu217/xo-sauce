
// {{ .Name }} returns ...
func {{ .Name }}({{ goparamlist .Params true }}) ({{ retype .Return }}, error) {
	res, err := {{ .Package }}.{{ .Name }}({{ goparamlist .Params false }})
	if err != nil {
		return {{ reniltype .Return }}, err
	}

	// convert package and return result
{{- if .Return.IsArray }}
	ress := make({{ retype .Return }}, len(res))
	for i, r := range res {
		ress[i] = &{{ .Return.Type }}{
		    {{ .Return.Type }}: r,
	    }
	}
	return ress, nil
{{ else }}
	return &{{ .Return.Type }}{
		{{ .Return.Type }}: res,
	}, nil
{{ end }}
}
