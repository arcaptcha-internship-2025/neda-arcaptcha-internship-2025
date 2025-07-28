package main

type Permissions struct { //from least to most significant bit
	canSeeMessages      bool
	canDeleteMessages   bool
	canEditMessages     bool
	canKickMembers      bool
	canMakeMembersAdmin bool
	canAddMembers       bool
}

func SetUserPermissions(permissions *Permissions) int8 {
	var mask int8 = 0
	//using bitwise or to set the appropriate bits
	if permissions.canSeeMessages {
		mask |= 1 << 0
	}
	if permissions.canDeleteMessages {
		mask |= 1 << 1
	}
	if permissions.canEditMessages {
		mask |= 1 << 2
	}
	if permissions.canKickMembers {
		mask |= 1 << 3
	}
	if permissions.canMakeMembersAdmin {
		mask |= 1 << 4
	}
	if permissions.canAddMembers {
		mask |= 1 << 5
	}

	return mask
}

func GetUserPermissions(permissionsMask int8) *Permissions {
	//taking int8 bitmask and converts it to  Permissions struct
	//then using bitwise and to check if each bit is set
	return &Permissions{
		canSeeMessages:      (permissionsMask & (1 << 0)) != 0,
		canDeleteMessages:   (permissionsMask & (1 << 1)) != 0,
		canEditMessages:     (permissionsMask & (1 << 2)) != 0,
		canKickMembers:      (permissionsMask & (1 << 3)) != 0,
		canMakeMembersAdmin: (permissionsMask & (1 << 4)) != 0,
		canAddMembers:       (permissionsMask & (1 << 5)) != 0,
	}
}
