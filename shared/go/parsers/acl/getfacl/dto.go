package getfacl

// Permission represents one rwx permission triplet.
type Permission struct {
	Read    bool
	Write   bool
	Execute bool
}

// Document contains parsed getfacl metadata and ACL entries for one path.
type Document struct {
	FilePath    string
	Owner       string
	Group       string
	User        Permission
	GroupObject Permission
	Other       Permission
	Mask        *Permission
	NamedUsers  map[string]Permission
	NamedGroups map[string]Permission
}
