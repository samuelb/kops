/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kops

import (
	"fmt"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const LabelClusterName = "kops.k8s.io/cluster"

// Deprecated - use the new labels & taints node-role.kubernetes.io/master and node-role.kubernetes.io/node
const TaintNoScheduleMaster15 = "dedicated=master:NoSchedule"

// InstanceGroup represents a group of instances (either nodes or masters) with the same configuration
type InstanceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstanceGroupSpec `json:"spec,omitempty"`
}

type InstanceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []InstanceGroup `json:"items"`
}

// InstanceGroupRole string describes the roles of the nodes in this InstanceGroup (master or nodes)
type InstanceGroupRole string

const (
	InstanceGroupRoleMaster  InstanceGroupRole = "Master"
	InstanceGroupRoleNode    InstanceGroupRole = "Node"
	InstanceGroupRoleBastion InstanceGroupRole = "Bastion"
)

var AllInstanceGroupRoles = []InstanceGroupRole{
	InstanceGroupRoleNode,
	InstanceGroupRoleMaster,
	InstanceGroupRoleBastion,
}

type InstanceGroupSpec struct {
	// Type determines the role of instances in this group: masters or nodes
	Role InstanceGroupRole `json:"role,omitempty"`

	Image   string `json:"image,omitempty"`
	MinSize *int32 `json:"minSize,omitempty"`
	MaxSize *int32 `json:"maxSize,omitempty"`
	//NodeInstancePrefix string `json:",omitempty"`
	//NodeLabels         string `json:",omitempty"`
	MachineType string `json:"machineType,omitempty"`
	//NodeTag            string `json:",omitempty"`

	// RootVolumeSize is the size of the EBS root volume to use, in GB
	RootVolumeSize *int32 `json:"rootVolumeSize,omitempty"`
	// RootVolumeType is the type of the EBS root volume to use (e.g. gp2)
	RootVolumeType *string `json:"rootVolumeType,omitempty"`

	Subnets []string `json:"subnets,omitempty"`

	// MaxPrice indicates this is a spot-pricing group, with the specified value as our max-price bid
	MaxPrice *string `json:"maxPrice,omitempty"`

	// AssociatePublicIP is true if we want instances to have a public IP
	AssociatePublicIP *bool `json:"associatePublicIp,omitempty"`

	// AdditionalSecurityGroups attaches additional security groups (e.g. i-123456)
	AdditionalSecurityGroups []string `json:"additionalSecurityGroups,omitempty"`

	// CloudLabels indicates the labels for instances in this group, at the AWS level
	CloudLabels map[string]string `json:"cloudLabels,omitempty"`

	// NodeLabels indicates the kubernetes labels for nodes in this group
	NodeLabels map[string]string `json:"nodeLabels,omitempty"`

	// Describes the tenancy of the instance group. Can be either default or dedicated.
	// Currently only applies to AWS.
	Tenancy string `json:"tenancy,omitempty"`

	// Kubelet overrides kubelet config from the ClusterSpec
	Kubelet *KubeletConfigSpec `json:"kubelet,omitempty"`

	// Taints indicates the kubernetes taints for nodes in this group
	Taints []string `json:"taints,omitempty"`
}

// PerformAssignmentsInstanceGroups populates InstanceGroups with default values
func PerformAssignmentsInstanceGroups(groups []*InstanceGroup) error {
	names := map[string]bool{}
	for _, group := range groups {
		names[group.ObjectMeta.Name] = true
	}

	for _, group := range groups {
		// We want to give them a stable Name as soon as possible
		if group.ObjectMeta.Name == "" {
			// Loop to find the first unassigned name like `nodes-%d`
			i := 0
			for {
				key := fmt.Sprintf("nodes-%d", i)
				if !names[key] {
					group.ObjectMeta.Name = key
					names[key] = true
					break
				}
				i++
			}
		}
	}

	return nil
}

func (g *InstanceGroup) IsMaster() bool {
	switch g.Spec.Role {
	case InstanceGroupRoleMaster:
		return true
	case InstanceGroupRoleNode:
		return false
	case InstanceGroupRoleBastion:
		return false

	default:
		glog.Fatalf("Role not set in group %v", g)
		return false
	}
}
