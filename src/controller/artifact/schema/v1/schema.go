package v1

const (
	// XHarborAttributes is the specific key defined in artifact config layer
	// for user-defined artifact to configure XHarborAttribute JSON schema
	XHarborAttributes = "xHarborAttributes"
	// SchemaVersionV1 1
	SchemaVersionV1 = 1
)

// Schema define a JSON schema for default processor to extract artifact layer data
// to artifact model in a specific way
//
// example:
// {
//        "schemaVersion": 1,
//        "icon": "https://github.com/caicloud/ormb/raw/master/docs/images/logo.png",
//        "additions": [
//            {
//                "contentType": "text/plain; charset=utf-8",
//                "name": "yaml",
//                "digest": "sha256:c2b304e60b7aec6a32d50b0d2c064933a7554db9d5d55259ac236f630a1c1f86"
//            },
//            {
//                "contentType": "text/plain; charset=utf-8",
//                "name": "readme",
//                "digest": "sha256:6dba1ad7ead7a5ee681441ec4b56b6a24690de6411d4574b927ce654c303f3c6"
//            }
//		  ],
//        "skipKeyList": [
//            "metrics",
//			  "dataset"
//        ]
// }
//
// For more details, please refer https://github.com/goharbor/community/master/proposals/assets/artifact-processor
//
// The schema JSON is under xHarborAttributes key in artifact config layer JSON
type Schema struct {
	// artifact config schema version
	SchemaVersion int `json:"schemaVersion"`
	// artifact icon online URL
	Icon string `json:"icon,omitempty"`
	// artifact additions
	Additions []AdditionSchema `json:"additions,omitempty"`
	// keys in SkipKeyList are used to specify which keys in artifact config JSON
	// will not be extracted out to artifact model stored in database
	SkipKeyList []string `json:"skipKeyList,omitempty"`
}

// AdditionSchema defines the specific addition of user-defined artifacts
type AdditionSchema struct {
	// the content type of the addition
	ContentType string `json:"contentType"`
	// addition name
	Name string `json:"name"`
	// addition layer digest
	Digest string `json:"digest"`
}
