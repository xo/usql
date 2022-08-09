package metadata

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAccessPrivileges_String(t *testing.T) {
	tests := []struct {
		name string
		ps   ObjectPrivileges
		want string
	}{
		{
			name: "multi",
			ps: ObjectPrivileges{
				{Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT", IsGrantable: true},
				{Grantee: "user1", Grantor: "user1", PrivilegeType: "SELECT"},
				{Grantee: "user2", Grantor: "user1", PrivilegeType: "INSERT"},
				{Grantee: "user2", Grantor: "user1", PrivilegeType: "SELECT", IsGrantable: true},
				{Grantee: "user3", Grantor: "user1", PrivilegeType: "SELECT", IsGrantable: true},
				{Grantee: "user3", Grantor: "user2", PrivilegeType: "UPDATE"},
			},
			want: "user1=INSERT*,SELECT/user1\n" +
				"user2=INSERT,SELECT*/user1\n" +
				"user3=SELECT*/user1\n" +
				"user3=UPDATE/user2",
		},
		{
			name: "one",
			ps: ObjectPrivileges{
				{Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT"},
			},
			want: "user1=INSERT/user1",
		},
		{
			name: "empty",
			ps:   ObjectPrivileges{},
			want: "",
		},
		{
			name: "empty-grantor",
			ps: ObjectPrivileges{
				{Grantee: "user1", Grantor: "", PrivilegeType: "INSERT", IsGrantable: true},
				{Grantee: "user1", Grantor: "", PrivilegeType: "SELECT"},
				{Grantee: "user2", Grantor: "", PrivilegeType: "INSERT"},
				{Grantee: "user2", Grantor: "", PrivilegeType: "SELECT", IsGrantable: true},
				{Grantee: "user3", Grantor: "", PrivilegeType: "UPDATE"},
			},
			want: "user1=INSERT*,SELECT\n" +
				"user2=INSERT,SELECT*\n" +
				"user3=UPDATE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ps.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Wrong AccessPrivileges.String(): (-expected, +got):\n%s", diff)
			}
		})
	}
}

func TestColumnPrivileges_String(t *testing.T) {
	tests := []struct {
		name string
		ps   ColumnPrivileges
		want string
	}{
		{
			name: "multi",
			ps: ColumnPrivileges{
				{Column: "col1", Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT", IsGrantable: true},
				{Column: "col1", Grantee: "user1", Grantor: "user1", PrivilegeType: "SELECT"},
				{Column: "col1", Grantee: "user2", Grantor: "user1", PrivilegeType: "INSERT"},
				{Column: "col1", Grantee: "user2", Grantor: "user1", PrivilegeType: "SELECT", IsGrantable: true},
				{Column: "col1", Grantee: "user3", Grantor: "user1", PrivilegeType: "SELECT", IsGrantable: true},
				{Column: "col1", Grantee: "user3", Grantor: "user2", PrivilegeType: "UPDATE"},
				{Column: "col2", Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT", IsGrantable: true},
				{Column: "col2", Grantee: "user1", Grantor: "user1", PrivilegeType: "SELECT"},
				{Column: "col2", Grantee: "user2", Grantor: "user1", PrivilegeType: "INSERT"},
				{Column: "col2", Grantee: "user2", Grantor: "user1", PrivilegeType: "SELECT", IsGrantable: true},
				{Column: "col2", Grantee: "user3", Grantor: "user2", PrivilegeType: "UPDATE"},
			},
			want: "col1:\n" +
				"  user1=INSERT*,SELECT/user1\n" +
				"  user2=INSERT,SELECT*/user1\n" +
				"  user3=SELECT*/user1\n" +
				"  user3=UPDATE/user2\n" +
				"col2:\n" +
				"  user1=INSERT*,SELECT/user1\n" +
				"  user2=INSERT,SELECT*/user1\n" +
				"  user3=UPDATE/user2",
		},
		{
			name: "one-multi",
			ps: ColumnPrivileges{
				{Column: "col2", Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT", IsGrantable: true},
				{Column: "col3", Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT", IsGrantable: true},
				{Column: "col3", Grantee: "user1", Grantor: "user1", PrivilegeType: "SELECT"},
				{Column: "col3", Grantee: "user2", Grantor: "user1", PrivilegeType: "INSERT"},
				{Column: "col3", Grantee: "user2", Grantor: "user1", PrivilegeType: "SELECT", IsGrantable: true},
				{Column: "col3", Grantee: "user3", Grantor: "user2", PrivilegeType: "UPDATE"},
			},
			want: "col2:\n" +
				"  user1=INSERT*/user1\n" +
				"col3:\n" +
				"  user1=INSERT*,SELECT/user1\n" +
				"  user2=INSERT,SELECT*/user1\n" +
				"  user3=UPDATE/user2",
		},
		{
			name: "multi-one",
			ps: ColumnPrivileges{
				{Column: "col1", Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT", IsGrantable: true},
				{Column: "col1", Grantee: "user1", Grantor: "user1", PrivilegeType: "SELECT"},
				{Column: "col1", Grantee: "user2", Grantor: "user1", PrivilegeType: "INSERT"},
				{Column: "col1", Grantee: "user2", Grantor: "user1", PrivilegeType: "SELECT", IsGrantable: true},
				{Column: "col1", Grantee: "user3", Grantor: "user2", PrivilegeType: "UPDATE"},
				{Column: "col2", Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT", IsGrantable: true},
			},
			want: "col1:\n" +
				"  user1=INSERT*,SELECT/user1\n" +
				"  user2=INSERT,SELECT*/user1\n" +
				"  user3=UPDATE/user2\n" +
				"col2:\n" +
				"  user1=INSERT*/user1",
		},
		{
			name: "one",
			ps: ColumnPrivileges{
				{Column: "col1", Grantee: "user1", Grantor: "user1", PrivilegeType: "INSERT"},
			},
			want: "col1:\n  user1=INSERT/user1",
		},
		{
			name: "empty",
			ps:   ColumnPrivileges{},
			want: "",
		},
		{
			name: "empty-grantor",
			ps: ColumnPrivileges{
				{Column: "col1", Grantee: "user1", Grantor: "", PrivilegeType: "INSERT", IsGrantable: true},
				{Column: "col1", Grantee: "user1", Grantor: "", PrivilegeType: "SELECT"},
				{Column: "col1", Grantee: "user2", Grantor: "", PrivilegeType: "INSERT"},
				{Column: "col1", Grantee: "user2", Grantor: "", PrivilegeType: "SELECT", IsGrantable: true},
				{Column: "col1", Grantee: "user3", Grantor: "", PrivilegeType: "UPDATE"},
				{Column: "col2", Grantee: "user1", Grantor: "", PrivilegeType: "INSERT", IsGrantable: true},
				{Column: "col2", Grantee: "user1", Grantor: "", PrivilegeType: "SELECT"},
				{Column: "col2", Grantee: "user2", Grantor: "", PrivilegeType: "INSERT"},
				{Column: "col2", Grantee: "user2", Grantor: "", PrivilegeType: "SELECT", IsGrantable: true},
				{Column: "col2", Grantee: "user3", Grantor: "", PrivilegeType: "UPDATE"},
			},
			want: "col1:\n" +
				"  user1=INSERT*,SELECT\n" +
				"  user2=INSERT,SELECT*\n" +
				"  user3=UPDATE\n" +
				"col2:\n" +
				"  user1=INSERT*,SELECT\n" +
				"  user2=INSERT,SELECT*\n" +
				"  user3=UPDATE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ps.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Wrong ColumnPrivileges.String(): (-expected, +got):\n%s", diff)
			}
		})
	}
}
