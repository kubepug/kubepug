package results

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// ListObjects is a generic function that receives unstructured Kubernetes objects and
// convert to deprecatedItem to be used later in the results
func ListObjects(items []unstructured.Unstructured) (deprecatedItems []Item) {
	for _, d := range items {
		name := d.GetName()
		namespace := d.GetNamespace()
		if namespace != "" {
			deprecatedItems = append(deprecatedItems, Item{Scope: "OBJECT", ObjectName: name, Namespace: namespace})
		} else {
			deprecatedItems = append(deprecatedItems, Item{Scope: "GLOBAL", ObjectName: name})
		}
	}

	return deprecatedItems
}

func CreateItem(group, version, kind string, items []Item) ResultItem {
	return ResultItem{
		Group:   group,
		Kind:    kind,
		Version: version,
		Items:   items,
	}
}
