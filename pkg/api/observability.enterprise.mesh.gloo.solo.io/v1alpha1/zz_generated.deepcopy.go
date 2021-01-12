// Code generated by skv2. DO NOT EDIT.

// This file contains generated Deepcopy methods for observability.enterprise.mesh.gloo.solo.io/v1alpha1 resources

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// Generated Deepcopy methods for AccessLogCollection

func (in *AccessLogCollection) DeepCopyInto(out *AccessLogCollection) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)

	// deepcopy spec
	in.Spec.DeepCopyInto(&out.Spec)
	// deepcopy status
	in.Status.DeepCopyInto(&out.Status)

	return
}

func (in *AccessLogCollection) DeepCopy() *AccessLogCollection {
	if in == nil {
		return nil
	}
	out := new(AccessLogCollection)
	in.DeepCopyInto(out)
	return out
}

func (in *AccessLogCollection) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *AccessLogCollectionList) DeepCopyInto(out *AccessLogCollectionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AccessLogCollection, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

func (in *AccessLogCollectionList) DeepCopy() *AccessLogCollectionList {
	if in == nil {
		return nil
	}
	out := new(AccessLogCollectionList)
	in.DeepCopyInto(out)
	return out
}

func (in *AccessLogCollectionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
