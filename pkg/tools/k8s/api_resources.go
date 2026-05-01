package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

type APIGroupDiscovery struct {
	Name             string
	Versions         []string
	PreferredVersion string
}

// GroupVersionResourceList is a wrapper for the API group, version, and resource list.
type GroupVersionResourceList struct {
	Group        *metav1.APIGroup
	Version      *metav1.GroupVersionForDiscovery
	ResourceList *metav1.APIResourceList
}

func processAPIResourceList(resourceList *metav1.APIResourceList, apiVersion string) []APIResourceInfo {
	var resourceInfos []APIResourceInfo
	for _, resource := range resourceList.APIResources {
		// Subresources are not explicitly listed in the api-resources output.
		if !strings.Contains(resource.Name, "/") {
			resourceInfo := NewAPIResourceInfoFromAPIResource(resource, apiVersion)
			resourceInfos = append(resourceInfos, resourceInfo)
		}
	}
	return resourceInfos
}

func listK8sAPIResources(ctx context.Context, _ *mcp.CallToolRequest, args *listK8sAPIResourcesArgs) (*mcp.CallToolResult, any, error) {
	_, client, res, err := getK8sDiscovery(&args.Cluster)
	if res != nil || err != nil {
		return res, nil, err
	}

	gvrList, err := listK8sAPIResourcesImpl(ctx, client)
	if err != nil {
		return textResult("Error getting API resources: %v", err), nil, nil
	}

	resourceVersions := make(map[string][]string)
	resourcePreferredVersion := make(map[string]string)
	for _, gvr := range gvrList {
		for _, resource := range gvr.ResourceList.APIResources {
			if strings.Contains(resource.Name, "/") {
				continue
			}
			resourceVersions[resource.Name] = append(resourceVersions[resource.Name], gvr.Version.GroupVersion)
			resourcePreferredVersion[resource.Name] = gvr.Group.PreferredVersion.GroupVersion
		}
	}

	var resources []APIGroupDiscovery
	for name, versions := range resourceVersions {
		resources = append(resources, APIGroupDiscovery{
			Name:             name,
			Versions:         versions,
			PreferredVersion: resourcePreferredVersion[name],
		})
	}

	bytes, err := json.Marshal(resources)
	if err != nil {
		return textResult("Error marshalling API resources: %v", err), nil, nil
	}

	return textResult("%s", string(bytes)), nil, nil
}

func listK8sAPIResourcesImpl(ctx context.Context, client discovery.DiscoveryInterface) ([]*GroupVersionResourceList, error) {
	// Get all API groups and resources using aggregated discovery.
	groups, apiResourceLists, err := client.ServerGroupsAndResources()
	if err != nil && len(apiResourceLists) == 0 {
		return nil, fmt.Errorf("failed to get server groups and resources: %w", err)
	}
	if err != nil {
		log.Printf("ServerGroupsAndResources returned partial results with error: %v", err)
	}

	// Map APIResourceList by GroupVersion for quick lookup.
	resourceListMap := make(map[string]*metav1.APIResourceList)
	for _, rl := range apiResourceLists {
		resourceListMap[rl.GroupVersion] = rl
	}

	var gvrList []*GroupVersionResourceList
	for _, group := range groups {
		for _, version := range group.Versions {
			v := version
			if rl, ok := resourceListMap[v.GroupVersion]; ok {
				gvrList = append(gvrList, &GroupVersionResourceList{
					Group:        group,
					Version:      &v,
					ResourceList: rl,
				})
			} else {
				log.Printf("Could not get resources for %s", v.GroupVersion)
			}
		}
	}

	return gvrList, nil
}
