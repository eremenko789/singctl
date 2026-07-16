package output

import (
	"encoding/json"
	"io"
)

func renderJSON(w io.Writer, set RecordSet, opts RenderOptions) error {
	layout := layoutOf(opts)
	data := mapsForJSONYAML(set, layout)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
