package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccResourceVSphereVmfsDatastore(t *testing.T) {
	var tp *testing.T
	testAccResourceVSphereVmfsDatastoreCases := []struct {
		name     string
		testCase resource.TestCase
	}{
		{
			"basic",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereVmfsDatastorePreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereVmfsDatastoreExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereVmfsDatastoreConfigStaticSingle(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereVmfsDatastoreExists(true),
						),
					},
				},
			},
		},
		{
			"multi-disk",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereVmfsDatastorePreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereVmfsDatastoreExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereVmfsDatastoreConfigStaticMulti(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereVmfsDatastoreExists(true),
						),
					},
				},
			},
		},
		{
			"discovery via data source",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereVmfsDatastorePreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereVmfsDatastoreExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereVmfsDatastoreConfigDiscoverDatasource(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereVmfsDatastoreExists(true),
						),
					},
				},
			},
		},
		{
			"add disks through update",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereVmfsDatastorePreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereVmfsDatastoreExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereVmfsDatastoreConfigStaticSingle(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereVmfsDatastoreExists(true),
						),
					},
					{
						Config: testAccResourceVSphereVmfsDatastoreConfigStaticMulti(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereVmfsDatastoreExists(true),
						),
					},
				},
			},
		},
		{
			"rename datastore",
			resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(tp)
					testAccResourceVSphereVmfsDatastorePreCheck(tp)
				},
				Providers:    testAccProviders,
				CheckDestroy: testAccResourceVSphereVmfsDatastoreExists(false),
				Steps: []resource.TestStep{
					{
						Config: testAccResourceVSphereVmfsDatastoreConfigStaticSingle(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereVmfsDatastoreExists(true),
						),
					},
					{
						Config: testAccResourceVSphereVmfsDatastoreConfigStaticSingleAltName(),
						Check: resource.ComposeTestCheckFunc(
							testAccResourceVSphereVmfsDatastoreExists(true),
							testAccResourceVSphereVmfsDatastoreHasName("terraform-test-renamed"),
						),
					},
				},
			},
		},
	}

	for _, tc := range testAccResourceVSphereVmfsDatastoreCases {
		t.Run(tc.name, func(t *testing.T) {
			tp = t
			resource.Test(t, tc.testCase)
		})
	}
}

func testAccResourceVSphereVmfsDatastorePreCheck(t *testing.T) {
	if os.Getenv("VSPHERE_ESXI_HOST") == "" {
		t.Skip("set VSPHERE_ESXI_HOST to run vsphere_vmfs_disks acceptance tests")
	}
	if os.Getenv("VSPHERE_DS_VMFS_DISK0") == "" {
		t.Skip("set VSPHERE_DS_VMFS_DISK0 to run vsphere_vmfs_datastore acceptance tests")
	}
	if os.Getenv("VSPHERE_DS_VMFS_DISK1") == "" {
		t.Skip("set VSPHERE_DS_VMFS_DISK1 to run vsphere_vmfs_datastore acceptance tests")
	}
	if os.Getenv("VSPHERE_DS_VMFS_DISK2") == "" {
		t.Skip("set VSPHERE_DS_VMFS_DISK2 to run vsphere_vmfs_datastore acceptance tests")
	}
	if os.Getenv("VSPHERE_VMFS_REGEXP") == "" {
		t.Skip("set VSPHERE_VMFS_REGEXP to run vsphere_vmfs_datastore acceptance tests")
	}
}

func testAccResourceVSphereVmfsDatastoreExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vars, err := testClientVariablesForResource(s, "vsphere_vmfs_datastore.datastore")
		if err != nil {
			return err
		}

		_, err = datastoreFromID(vars.client, vars.resourceID)
		if err != nil {
			if isManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected datastore %s to be missing", vars.resourceID)
		}
		return nil
	}
}

func testAccResourceVSphereVmfsDatastoreHasName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vars, err := testClientVariablesForResource(s, "vsphere_vmfs_datastore.datastore")
		if err != nil {
			return err
		}

		ds, err := datastoreFromID(vars.client, vars.resourceID)
		if err != nil {
			return err
		}

		props, err := datastoreProperties(ds)
		if err != nil {
			return err
		}

		actual := props.Summary.Name
		if expected != actual {
			return fmt.Errorf("expected datastore name to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereVmfsDatastoreConfigStaticSingle() string {
	return fmt.Sprintf(`
variable "disk0" {
  type    = "string"
  default = "%s"
}

data "vsphere_datacenter" "datacenter" {
  name = "%s"
}

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.datacenter.id}"
}

resource "vsphere_vmfs_datastore" "datastore" {
  name           = "terraform-test"
  host_system_id = "${data.vsphere_host.esxi_host.id}"

  disks = [
    "${var.disk0}",
  ]
}
`, os.Getenv("VSPHERE_DS_VMFS_DISK0"), os.Getenv("VSPHERE_DATACENTER"), os.Getenv("VSPHERE_ESXI_HOST"))
}

func testAccResourceVSphereVmfsDatastoreConfigStaticSingleAltName() string {
	return fmt.Sprintf(`
variable "disk0" {
  type    = "string"
  default = "%s"
}

data "vsphere_datacenter" "datacenter" {
  name = "%s"
}

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.datacenter.id}"
}

resource "vsphere_vmfs_datastore" "datastore" {
  name           = "terraform-test-renamed"
  host_system_id = "${data.vsphere_host.esxi_host.id}"

  disks = [
    "${var.disk0}",
  ]
}
`, os.Getenv("VSPHERE_DS_VMFS_DISK0"), os.Getenv("VSPHERE_DATACENTER"), os.Getenv("VSPHERE_ESXI_HOST"))
}

func testAccResourceVSphereVmfsDatastoreConfigStaticMulti() string {
	return fmt.Sprintf(`
variable "disk0" {
  type    = "string"
  default = "%s"
}

variable "disk1" {
  type    = "string"
  default = "%s"
}

variable "disk2" {
  type    = "string"
  default = "%s"
}

data "vsphere_datacenter" "datacenter" {
  name = "%s"
}

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.datacenter.id}"
}

resource "vsphere_vmfs_datastore" "datastore" {
  name           = "terraform-test"
  host_system_id = "${data.vsphere_host.esxi_host.id}"

  disks = [
    "${var.disk0}",
    "${var.disk1}",
    "${var.disk2}",
  ]
}
`, os.Getenv("VSPHERE_DS_VMFS_DISK0"), os.Getenv("VSPHERE_DS_VMFS_DISK1"), os.Getenv("VSPHERE_DS_VMFS_DISK2"), os.Getenv("VSPHERE_DATACENTER"), os.Getenv("VSPHERE_ESXI_HOST"))
}

func testAccResourceVSphereVmfsDatastoreConfigDiscoverDatasource() string {
	return fmt.Sprintf(`
variable "regexp" {
  type    = "string"
  default = "%s"
}

data "vsphere_datacenter" "datacenter" {
  name = "%s"
}

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.datacenter.id}"
}

data "vsphere_vmfs_disks" "available" {
  host_system_id = "${data.vsphere_host.esxi_host.id}"
  rescan         = true
  filter         = "${var.regexp}"
}

resource "vsphere_vmfs_datastore" "datastore" {
  name           = "terraform-test"
  host_system_id = "${data.vsphere_host.esxi_host.id}"

  disks = ["${data.vsphere_vmfs_disks.available.disks}"]
}
`, os.Getenv("VSPHERE_VMFS_REGEXP"), os.Getenv("VSPHERE_DATACENTER"), os.Getenv("VSPHERE_ESXI_HOST"))
}
