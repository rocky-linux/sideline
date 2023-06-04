// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.3
// source: configuration.proto

package sidelinepb

import (
	pb "github.com/rocky-linux/srpmproc/pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Upstream_CompareMode int32

const (
	Upstream_Unknown Upstream_CompareMode = 0
	// Reduce refspecs we fetch for the kernel
	// TM = Target major version, CV = Compare version
	// Uses following refspecs:
	//   * refs/tags/vTM.*
	//   * refs/tags/vCV
	Upstream_KernelTag Upstream_CompareMode = 1
)

// Enum value maps for Upstream_CompareMode.
var (
	Upstream_CompareMode_name = map[int32]string{
		0: "Unknown",
		1: "KernelTag",
	}
	Upstream_CompareMode_value = map[string]int32{
		"Unknown":   0,
		"KernelTag": 1,
	}
)

func (x Upstream_CompareMode) Enum() *Upstream_CompareMode {
	p := new(Upstream_CompareMode)
	*p = x
	return p
}

func (x Upstream_CompareMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Upstream_CompareMode) Descriptor() protoreflect.EnumDescriptor {
	return file_configuration_proto_enumTypes[0].Descriptor()
}

func (Upstream_CompareMode) Type() protoreflect.EnumType {
	return &file_configuration_proto_enumTypes[0]
}

func (x Upstream_CompareMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Upstream_CompareMode.Descriptor instead.
func (Upstream_CompareMode) EnumDescriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{0, 0}
}

type Upstream struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Target forge
	//
	// Types that are assignable to Forge:
	//	*Upstream_Git
	Forge isUpstream_Forge `protobuf_oneof:"forge"`
	// Depth we need to clone
	// For tags or branches targeting latest, only 1 is needed for example
	Depth int32 `protobuf:"varint,2,opt,name=depth,proto3" json:"depth,omitempty"`
	// Target revision
	//
	// Types that are assignable to Target:
	//	*Upstream_Tag
	//	*Upstream_Branch
	Target isUpstream_Target `protobuf_oneof:"target"`
	// Compare revision (optional)
	//
	// Types that are assignable to Compare:
	//	*Upstream_CompareWithTag
	Compare isUpstream_Compare `protobuf_oneof:"compare"`
	// Whether the compare repo should be laid out in disk
	// Currently always true whether specified or not because
	// of how `git log` operation is implemented
	CompareShouldUseOsfs bool                 `protobuf:"varint,5,opt,name=compare_should_use_osfs,json=compareShouldUseOsfs,proto3" json:"compare_should_use_osfs,omitempty"`
	CompareMode          Upstream_CompareMode `protobuf:"varint,6,opt,name=compare_mode,json=compareMode,proto3,enum=sideline.Upstream_CompareMode" json:"compare_mode,omitempty"`
}

func (x *Upstream) Reset() {
	*x = Upstream{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configuration_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Upstream) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Upstream) ProtoMessage() {}

func (x *Upstream) ProtoReflect() protoreflect.Message {
	mi := &file_configuration_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Upstream.ProtoReflect.Descriptor instead.
func (*Upstream) Descriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{0}
}

func (m *Upstream) GetForge() isUpstream_Forge {
	if m != nil {
		return m.Forge
	}
	return nil
}

func (x *Upstream) GetGit() string {
	if x, ok := x.GetForge().(*Upstream_Git); ok {
		return x.Git
	}
	return ""
}

func (x *Upstream) GetDepth() int32 {
	if x != nil {
		return x.Depth
	}
	return 0
}

func (m *Upstream) GetTarget() isUpstream_Target {
	if m != nil {
		return m.Target
	}
	return nil
}

func (x *Upstream) GetTag() string {
	if x, ok := x.GetTarget().(*Upstream_Tag); ok {
		return x.Tag
	}
	return ""
}

func (x *Upstream) GetBranch() string {
	if x, ok := x.GetTarget().(*Upstream_Branch); ok {
		return x.Branch
	}
	return ""
}

func (m *Upstream) GetCompare() isUpstream_Compare {
	if m != nil {
		return m.Compare
	}
	return nil
}

func (x *Upstream) GetCompareWithTag() *wrapperspb.StringValue {
	if x, ok := x.GetCompare().(*Upstream_CompareWithTag); ok {
		return x.CompareWithTag
	}
	return nil
}

func (x *Upstream) GetCompareShouldUseOsfs() bool {
	if x != nil {
		return x.CompareShouldUseOsfs
	}
	return false
}

func (x *Upstream) GetCompareMode() Upstream_CompareMode {
	if x != nil {
		return x.CompareMode
	}
	return Upstream_Unknown
}

type isUpstream_Forge interface {
	isUpstream_Forge()
}

type Upstream_Git struct {
	// A git remote
	Git string `protobuf:"bytes,1,opt,name=git,proto3,oneof"`
}

func (*Upstream_Git) isUpstream_Forge() {}

type isUpstream_Target interface {
	isUpstream_Target()
}

type Upstream_Tag struct {
	// Caller wants a tag
	Tag string `protobuf:"bytes,3,opt,name=tag,proto3,oneof"`
}

type Upstream_Branch struct {
	Branch string `protobuf:"bytes,7,opt,name=branch,proto3,oneof"`
}

func (*Upstream_Tag) isUpstream_Target() {}

func (*Upstream_Branch) isUpstream_Target() {}

type isUpstream_Compare interface {
	isUpstream_Compare()
}

type Upstream_CompareWithTag struct {
	CompareWithTag *wrapperspb.StringValue `protobuf:"bytes,4,opt,name=compare_with_tag,json=compareWithTag,proto3,oneof"`
}

func (*Upstream_CompareWithTag) isUpstream_Compare() {}

type SearchReplace struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Find    string `protobuf:"bytes,1,opt,name=find,proto3" json:"find,omitempty"`
	Replace string `protobuf:"bytes,2,opt,name=replace,proto3" json:"replace,omitempty"`
	N       int32  `protobuf:"varint,3,opt,name=n,proto3" json:"n,omitempty"`
}

func (x *SearchReplace) Reset() {
	*x = SearchReplace{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configuration_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchReplace) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchReplace) ProtoMessage() {}

func (x *SearchReplace) ProtoReflect() protoreflect.Message {
	mi := &file_configuration_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchReplace.ProtoReflect.Descriptor instead.
func (*SearchReplace) Descriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{1}
}

func (x *SearchReplace) GetFind() string {
	if x != nil {
		return x.Find
	}
	return ""
}

func (x *SearchReplace) GetReplace() string {
	if x != nil {
		return x.Replace
	}
	return ""
}

func (x *SearchReplace) GetN() int32 {
	if x != nil {
		return x.N
	}
	return 0
}

type FileChange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Path to file
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Search and replace operations on file
	SearchReplace []*SearchReplace `protobuf:"bytes,2,rep,name=search_replace,json=searchReplace,proto3" json:"search_replace,omitempty"`
}

func (x *FileChange) Reset() {
	*x = FileChange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configuration_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileChange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileChange) ProtoMessage() {}

func (x *FileChange) ProtoReflect() protoreflect.Message {
	mi := &file_configuration_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileChange.ProtoReflect.Descriptor instead.
func (*FileChange) Descriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{2}
}

func (x *FileChange) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *FileChange) GetSearchReplace() []*SearchReplace {
	if x != nil {
		return x.SearchReplace
	}
	return nil
}

type Changes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Paths listed here are recursively copied into Source0
	// and overridden (always choosing upstream)
	RecursivePath []string      `protobuf:"bytes,1,rep,name=recursive_path,json=recursivePath,proto3" json:"recursive_path,omitempty"`
	FileChange    []*FileChange `protobuf:"bytes,2,rep,name=file_change,json=fileChange,proto3" json:"file_change,omitempty"`
}

func (x *Changes) Reset() {
	*x = Changes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configuration_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Changes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Changes) ProtoMessage() {}

func (x *Changes) ProtoReflect() protoreflect.Message {
	mi := &file_configuration_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Changes.ProtoReflect.Descriptor instead.
func (*Changes) Descriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{3}
}

func (x *Changes) GetRecursivePath() []string {
	if x != nil {
		return x.RecursivePath
	}
	return nil
}

func (x *Changes) GetFileChange() []*FileChange {
	if x != nil {
		return x.FileChange
	}
	return nil
}

type PresetOverride struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Override:
	//	*PresetOverride_BaseUrl
	//	*PresetOverride_FullUrl
	Override isPresetOverride_Override `protobuf_oneof:"override"`
}

func (x *PresetOverride) Reset() {
	*x = PresetOverride{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configuration_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PresetOverride) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PresetOverride) ProtoMessage() {}

func (x *PresetOverride) ProtoReflect() protoreflect.Message {
	mi := &file_configuration_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PresetOverride.ProtoReflect.Descriptor instead.
func (*PresetOverride) Descriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{4}
}

func (m *PresetOverride) GetOverride() isPresetOverride_Override {
	if m != nil {
		return m.Override
	}
	return nil
}

func (x *PresetOverride) GetBaseUrl() string {
	if x, ok := x.GetOverride().(*PresetOverride_BaseUrl); ok {
		return x.BaseUrl
	}
	return ""
}

func (x *PresetOverride) GetFullUrl() string {
	if x, ok := x.GetOverride().(*PresetOverride_FullUrl); ok {
		return x.FullUrl
	}
	return ""
}

type isPresetOverride_Override interface {
	isPresetOverride_Override()
}

type PresetOverride_BaseUrl struct {
	BaseUrl string `protobuf:"bytes,1,opt,name=base_url,json=baseUrl,proto3,oneof"`
}

type PresetOverride_FullUrl struct {
	FullUrl string `protobuf:"bytes,2,opt,name=full_url,json=fullUrl,proto3,oneof"`
}

func (*PresetOverride_BaseUrl) isPresetOverride_Override() {}

func (*PresetOverride_FullUrl) isPresetOverride_Override() {}

type AutoPatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Whether to add the patch to prep or not
	// Packages that are not auto-patched will need to set this and n_path
	AddToPrep bool `protobuf:"varint,1,opt,name=add_to_prep,json=addToPrep,proto3" json:"add_to_prep,omitempty"`
	// Patch strip depth
	NPath int32 `protobuf:"varint,2,opt,name=n_path,json=nPath,proto3" json:"n_path,omitempty"`
}

func (x *AutoPatch) Reset() {
	*x = AutoPatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configuration_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AutoPatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AutoPatch) ProtoMessage() {}

func (x *AutoPatch) ProtoReflect() protoreflect.Message {
	mi := &file_configuration_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AutoPatch.ProtoReflect.Descriptor instead.
func (*AutoPatch) Descriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{5}
}

func (x *AutoPatch) GetAddToPrep() bool {
	if x != nil {
		return x.AddToPrep
	}
	return false
}

func (x *AutoPatch) GetNPath() int32 {
	if x != nil {
		return x.NPath
	}
	return 0
}

type ApplyPatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Patch:
	//	*ApplyPatch_AutoPatch
	//	*ApplyPatch_Custom
	Patch isApplyPatch_Patch `protobuf_oneof:"patch"`
}

func (x *ApplyPatch) Reset() {
	*x = ApplyPatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configuration_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ApplyPatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ApplyPatch) ProtoMessage() {}

func (x *ApplyPatch) ProtoReflect() protoreflect.Message {
	mi := &file_configuration_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ApplyPatch.ProtoReflect.Descriptor instead.
func (*ApplyPatch) Descriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{6}
}

func (m *ApplyPatch) GetPatch() isApplyPatch_Patch {
	if m != nil {
		return m.Patch
	}
	return nil
}

func (x *ApplyPatch) GetAutoPatch() *AutoPatch {
	if x, ok := x.GetPatch().(*ApplyPatch_AutoPatch); ok {
		return x.AutoPatch
	}
	return nil
}

func (x *ApplyPatch) GetCustom() *pb.SpecChange {
	if x, ok := x.GetPatch().(*ApplyPatch_Custom); ok {
		return x.Custom
	}
	return nil
}

type isApplyPatch_Patch interface {
	isApplyPatch_Patch()
}

type ApplyPatch_AutoPatch struct {
	// Automatically applied patch (may be problematic with complex specs)
	// Ex. kernel should mostly use custom
	AutoPatch *AutoPatch `protobuf:"bytes,1,opt,name=auto_patch,json=autoPatch,proto3,oneof"`
}

type ApplyPatch_Custom struct {
	// Custom spec change command for this patch (excluding add patch)
	Custom *pb.SpecChange `protobuf:"bytes,2,opt,name=custom,proto3,oneof"`
}

func (*ApplyPatch_AutoPatch) isApplyPatch_Patch() {}

func (*ApplyPatch_Custom) isApplyPatch_Patch() {}

type Configuration struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Preset for source package
	// Currently only supports "rocky8" or "manual"
	// "rocky8" sets "base_url" to "https://dl.rockylinux.org/pub/rocky/8"
	// "manual" sets nothing
	Preset string `protobuf:"bytes,1,opt,name=preset,proto3" json:"preset,omitempty"`
	// Source package name
	// SRPM to be fetched from the source repository
	// Format: {repo}//{name}
	// Example: BaseOS//kernel
	Package string `protobuf:"bytes,2,opt,name=package,proto3" json:"package,omitempty"`
	// Upstream information
	// Where the upstream of this package is located
	// Only git is supported for now
	Upstream *Upstream `protobuf:"bytes,3,opt,name=upstream,proto3" json:"upstream,omitempty"`
	// Changes we want to pull from upstream
	Changes []*Changes `protobuf:"bytes,4,rep,name=changes,proto3" json:"changes,omitempty"`
	// Manual override for preset
	// Only used if preset is "manual"
	PresetOverride *PresetOverride `protobuf:"bytes,5,opt,name=preset_override,json=presetOverride,proto3" json:"preset_override,omitempty"`
	// How to apply the patch to the spec
	ApplyPatch *ApplyPatch `protobuf:"bytes,6,opt,name=apply_patch,json=applyPatch,proto3" json:"apply_patch,omitempty"`
	// Contact information
	ContactEmail string `protobuf:"bytes,7,opt,name=contact_email,json=contactEmail,proto3" json:"contact_email,omitempty"`
	ContactName  string `protobuf:"bytes,8,opt,name=contact_name,json=contactName,proto3" json:"contact_name,omitempty"`
}

func (x *Configuration) Reset() {
	*x = Configuration{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configuration_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Configuration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Configuration) ProtoMessage() {}

func (x *Configuration) ProtoReflect() protoreflect.Message {
	mi := &file_configuration_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Configuration.ProtoReflect.Descriptor instead.
func (*Configuration) Descriptor() ([]byte, []int) {
	return file_configuration_proto_rawDescGZIP(), []int{7}
}

func (x *Configuration) GetPreset() string {
	if x != nil {
		return x.Preset
	}
	return ""
}

func (x *Configuration) GetPackage() string {
	if x != nil {
		return x.Package
	}
	return ""
}

func (x *Configuration) GetUpstream() *Upstream {
	if x != nil {
		return x.Upstream
	}
	return nil
}

func (x *Configuration) GetChanges() []*Changes {
	if x != nil {
		return x.Changes
	}
	return nil
}

func (x *Configuration) GetPresetOverride() *PresetOverride {
	if x != nil {
		return x.PresetOverride
	}
	return nil
}

func (x *Configuration) GetApplyPatch() *ApplyPatch {
	if x != nil {
		return x.ApplyPatch
	}
	return nil
}

func (x *Configuration) GetContactEmail() string {
	if x != nil {
		return x.ContactEmail
	}
	return ""
}

func (x *Configuration) GetContactName() string {
	if x != nil {
		return x.ContactName
	}
	return ""
}

var File_configuration_proto protoreflect.FileDescriptor

var file_configuration_proto_rawDesc = []byte{
	0x0a, 0x13, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x73, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x1a,
	0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x3b, 0x74, 0x68, 0x69, 0x72, 0x64, 0x5f, 0x70, 0x61, 0x72, 0x74, 0x79, 0x2f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x6f, 0x63, 0x6b, 0x79, 0x2d, 0x6c, 0x69,
	0x6e, 0x75, 0x78, 0x2f, 0x73, 0x72, 0x70, 0x6d, 0x70, 0x72, 0x6f, 0x63, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x63, 0x66, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xef, 0x02, 0x0a,
	0x08, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x12, 0x0a, 0x03, 0x67, 0x69, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x03, 0x67, 0x69, 0x74, 0x12, 0x14, 0x0a,
	0x05, 0x64, 0x65, 0x70, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x64, 0x65,
	0x70, 0x74, 0x68, 0x12, 0x12, 0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x01, 0x52, 0x03, 0x74, 0x61, 0x67, 0x12, 0x18, 0x0a, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63,
	0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63,
	0x68, 0x12, 0x48, 0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x72, 0x65, 0x5f, 0x77, 0x69, 0x74,
	0x68, 0x5f, 0x74, 0x61, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x02, 0x52, 0x0e, 0x63, 0x6f, 0x6d,
	0x70, 0x61, 0x72, 0x65, 0x57, 0x69, 0x74, 0x68, 0x54, 0x61, 0x67, 0x12, 0x35, 0x0a, 0x17, 0x63,
	0x6f, 0x6d, 0x70, 0x61, 0x72, 0x65, 0x5f, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x5f, 0x75, 0x73,
	0x65, 0x5f, 0x6f, 0x73, 0x66, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x14, 0x63, 0x6f,
	0x6d, 0x70, 0x61, 0x72, 0x65, 0x53, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x55, 0x73, 0x65, 0x4f, 0x73,
	0x66, 0x73, 0x12, 0x41, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x72, 0x65, 0x5f, 0x6d, 0x6f,
	0x64, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x73, 0x69, 0x64, 0x65, 0x6c,
	0x69, 0x6e, 0x65, 0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x43, 0x6f, 0x6d,
	0x70, 0x61, 0x72, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x72,
	0x65, 0x4d, 0x6f, 0x64, 0x65, 0x22, 0x29, 0x0a, 0x0b, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x72, 0x65,
	0x4d, 0x6f, 0x64, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10,
	0x00, 0x12, 0x0d, 0x0a, 0x09, 0x4b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x54, 0x61, 0x67, 0x10, 0x01,
	0x42, 0x07, 0x0a, 0x05, 0x66, 0x6f, 0x72, 0x67, 0x65, 0x42, 0x08, 0x0a, 0x06, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x42, 0x09, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x72, 0x65, 0x22, 0x4b,
	0x0a, 0x0d, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x66, 0x69, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66,
	0x69, 0x6e, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x12, 0x0c, 0x0a,
	0x01, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x6e, 0x22, 0x60, 0x0a, 0x0a, 0x46,
	0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74,
	0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x3e, 0x0a,
	0x0e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x5f, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x73, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x52, 0x0d,
	0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x22, 0x67, 0x0a,
	0x07, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x65, 0x63, 0x75,
	0x72, 0x73, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x0d, 0x72, 0x65, 0x63, 0x75, 0x72, 0x73, 0x69, 0x76, 0x65, 0x50, 0x61, 0x74, 0x68, 0x12,
	0x35, 0x0a, 0x0b, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x73, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e,
	0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x0a, 0x66, 0x69, 0x6c, 0x65,
	0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x22, 0x56, 0x0a, 0x0e, 0x50, 0x72, 0x65, 0x73, 0x65, 0x74,
	0x4f, 0x76, 0x65, 0x72, 0x72, 0x69, 0x64, 0x65, 0x12, 0x1b, 0x0a, 0x08, 0x62, 0x61, 0x73, 0x65,
	0x5f, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x62, 0x61,
	0x73, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x1b, 0x0a, 0x08, 0x66, 0x75, 0x6c, 0x6c, 0x5f, 0x75, 0x72,
	0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x66, 0x75, 0x6c, 0x6c, 0x55,
	0x72, 0x6c, 0x42, 0x0a, 0x0a, 0x08, 0x6f, 0x76, 0x65, 0x72, 0x72, 0x69, 0x64, 0x65, 0x22, 0x42,
	0x0a, 0x09, 0x41, 0x75, 0x74, 0x6f, 0x50, 0x61, 0x74, 0x63, 0x68, 0x12, 0x1e, 0x0a, 0x0b, 0x61,
	0x64, 0x64, 0x5f, 0x74, 0x6f, 0x5f, 0x70, 0x72, 0x65, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x09, 0x61, 0x64, 0x64, 0x54, 0x6f, 0x50, 0x72, 0x65, 0x70, 0x12, 0x15, 0x0a, 0x06, 0x6e,
	0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6e, 0x50, 0x61,
	0x74, 0x68, 0x22, 0x7b, 0x0a, 0x0a, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x50, 0x61, 0x74, 0x63, 0x68,
	0x12, 0x34, 0x0a, 0x0a, 0x61, 0x75, 0x74, 0x6f, 0x5f, 0x70, 0x61, 0x74, 0x63, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x73, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e,
	0x41, 0x75, 0x74, 0x6f, 0x50, 0x61, 0x74, 0x63, 0x68, 0x48, 0x00, 0x52, 0x09, 0x61, 0x75, 0x74,
	0x6f, 0x50, 0x61, 0x74, 0x63, 0x68, 0x12, 0x2e, 0x0a, 0x06, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x73, 0x72, 0x70, 0x6d, 0x70, 0x72, 0x6f,
	0x63, 0x2e, 0x53, 0x70, 0x65, 0x63, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x48, 0x00, 0x52, 0x06,
	0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x42, 0x07, 0x0a, 0x05, 0x70, 0x61, 0x74, 0x63, 0x68, 0x22,
	0xe0, 0x02, 0x0a, 0x0d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x73, 0x65, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x63,
	0x6b, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x12, 0x2e, 0x0a, 0x08, 0x75, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x2e, 0x55, 0x70, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x08, 0x75, 0x70, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x12, 0x2b, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e,
	0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x52, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73,
	0x12, 0x41, 0x0a, 0x0f, 0x70, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x6f, 0x76, 0x65, 0x72, 0x72,
	0x69, 0x64, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x73, 0x69, 0x64, 0x65,
	0x6c, 0x69, 0x6e, 0x65, 0x2e, 0x50, 0x72, 0x65, 0x73, 0x65, 0x74, 0x4f, 0x76, 0x65, 0x72, 0x72,
	0x69, 0x64, 0x65, 0x52, 0x0e, 0x70, 0x72, 0x65, 0x73, 0x65, 0x74, 0x4f, 0x76, 0x65, 0x72, 0x72,
	0x69, 0x64, 0x65, 0x12, 0x35, 0x0a, 0x0b, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x5f, 0x70, 0x61, 0x74,
	0x63, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x73, 0x69, 0x64, 0x65, 0x6c,
	0x69, 0x6e, 0x65, 0x2e, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x50, 0x61, 0x74, 0x63, 0x68, 0x52, 0x0a,
	0x61, 0x70, 0x70, 0x6c, 0x79, 0x50, 0x61, 0x74, 0x63, 0x68, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x63, 0x74, 0x5f, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x12,
	0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x4e, 0x61,
	0x6d, 0x65, 0x42, 0x2f, 0x5a, 0x2d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x72, 0x6f, 0x63, 0x6b, 0x79, 0x2d, 0x6c, 0x69, 0x6e, 0x75, 0x78, 0x2f, 0x73, 0x69, 0x64,
	0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2f, 0x70, 0x62, 0x3b, 0x73, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_configuration_proto_rawDescOnce sync.Once
	file_configuration_proto_rawDescData = file_configuration_proto_rawDesc
)

func file_configuration_proto_rawDescGZIP() []byte {
	file_configuration_proto_rawDescOnce.Do(func() {
		file_configuration_proto_rawDescData = protoimpl.X.CompressGZIP(file_configuration_proto_rawDescData)
	})
	return file_configuration_proto_rawDescData
}

var file_configuration_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_configuration_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_configuration_proto_goTypes = []interface{}{
	(Upstream_CompareMode)(0),      // 0: sideline.Upstream.CompareMode
	(*Upstream)(nil),               // 1: sideline.Upstream
	(*SearchReplace)(nil),          // 2: sideline.SearchReplace
	(*FileChange)(nil),             // 3: sideline.FileChange
	(*Changes)(nil),                // 4: sideline.Changes
	(*PresetOverride)(nil),         // 5: sideline.PresetOverride
	(*AutoPatch)(nil),              // 6: sideline.AutoPatch
	(*ApplyPatch)(nil),             // 7: sideline.ApplyPatch
	(*Configuration)(nil),          // 8: sideline.Configuration
	(*wrapperspb.StringValue)(nil), // 9: google.protobuf.StringValue
	(*pb.SpecChange)(nil),          // 10: srpmproc.SpecChange
}
var file_configuration_proto_depIdxs = []int32{
	9,  // 0: sideline.Upstream.compare_with_tag:type_name -> google.protobuf.StringValue
	0,  // 1: sideline.Upstream.compare_mode:type_name -> sideline.Upstream.CompareMode
	2,  // 2: sideline.FileChange.search_replace:type_name -> sideline.SearchReplace
	3,  // 3: sideline.Changes.file_change:type_name -> sideline.FileChange
	6,  // 4: sideline.ApplyPatch.auto_patch:type_name -> sideline.AutoPatch
	10, // 5: sideline.ApplyPatch.custom:type_name -> srpmproc.SpecChange
	1,  // 6: sideline.Configuration.upstream:type_name -> sideline.Upstream
	4,  // 7: sideline.Configuration.changes:type_name -> sideline.Changes
	5,  // 8: sideline.Configuration.preset_override:type_name -> sideline.PresetOverride
	7,  // 9: sideline.Configuration.apply_patch:type_name -> sideline.ApplyPatch
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_configuration_proto_init() }
func file_configuration_proto_init() {
	if File_configuration_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_configuration_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Upstream); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_configuration_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchReplace); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_configuration_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileChange); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_configuration_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Changes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_configuration_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PresetOverride); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_configuration_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AutoPatch); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_configuration_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ApplyPatch); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_configuration_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Configuration); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_configuration_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Upstream_Git)(nil),
		(*Upstream_Tag)(nil),
		(*Upstream_Branch)(nil),
		(*Upstream_CompareWithTag)(nil),
	}
	file_configuration_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*PresetOverride_BaseUrl)(nil),
		(*PresetOverride_FullUrl)(nil),
	}
	file_configuration_proto_msgTypes[6].OneofWrappers = []interface{}{
		(*ApplyPatch_AutoPatch)(nil),
		(*ApplyPatch_Custom)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_configuration_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_configuration_proto_goTypes,
		DependencyIndexes: file_configuration_proto_depIdxs,
		EnumInfos:         file_configuration_proto_enumTypes,
		MessageInfos:      file_configuration_proto_msgTypes,
	}.Build()
	File_configuration_proto = out.File
	file_configuration_proto_rawDesc = nil
	file_configuration_proto_goTypes = nil
	file_configuration_proto_depIdxs = nil
}
