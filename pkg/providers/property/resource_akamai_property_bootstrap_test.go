package property

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v8/pkg/iam"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v8/pkg/papi"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/str"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type testDataForPropertyBootstrap struct {
	propertyID      string
	name            string
	groupID         string
	contractID      string
	productID       string
	withoutPrefixes bool
}

func TestBootstrapResourceCreate(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		configPath string
		init       func(*testing.T, *papi.Mock, testDataForPropertyBootstrap)
		mockData   testDataForPropertyBootstrap
		error      *regexp.Regexp
	}{
		"create": {
			configPath: "testdata/TestResPropertyBootstrap/create.tf",
			init: func(t *testing.T, m *papi.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(m, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(m, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(m, data.propertyID, data.contractID, data.groupID)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
		},
		"create without prefixes": {
			configPath: "testdata/TestResPropertyBootstrap/create_without_prefixes.tf",
			init: func(t *testing.T, m *papi.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(m, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(m, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(m, data.propertyID, data.contractID, data.groupID)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID:      "prp_123",
				name:            "property_name",
				groupID:         "grp_1",
				contractID:      "ctr_2",
				productID:       "prd_3",
				withoutPrefixes: true,
			},
		},
		"create with interpretCreate error - group not found": {
			configPath: "testdata/TestResPropertyBootstrap/create.tf",
			init: func(t *testing.T, m *papi.Mock, data testDataForPropertyBootstrap) {
				req := papi.CreatePropertyRequest{
					GroupID:    data.groupID,
					ContractID: data.contractID,
					Property: papi.PropertyCreate{
						ProductID:    data.productID,
						PropertyName: data.name,
					},
				}
				m.On("CreateProperty", AnyCTX, req).Return(nil, fmt.Errorf(
					"%s: %w: %s", papi.ErrCreateProperty, papi.ErrNotFound, "not found")).Once()
				// mock empty groups - no group has been found, hence the expected error
				m.On("GetGroups", AnyCTX).Return(&papi.GetGroupsResponse{
					Groups: papi.GroupItems{
						Items: []*papi.Group{},
					},
				}, nil).Once()
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			error: regexp.MustCompile(`Error: group not found: grp_1`),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := &papi.Mock{}
			if test.init != nil {
				test.init(t, m, test.mockData)
			}

			useClient(m, nil, func() {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
					IsUnitTest:               true,
					Steps: []resource.TestStep{
						{
							Config:      testutils.LoadFixtureString(t, test.configPath),
							Check:       checkPropertyBootstrapAttributes(test.mockData),
							ExpectError: test.error,
						},
					},
				})
			})

			m.AssertExpectations(t)
		})
	}
}

func TestBootstrapResourceUpdate(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		configPathForCreate string
		configPathForUpdate string
		init                func(*testing.T, *papi.Mock, *iam.Mock, testDataForPropertyBootstrap)
		mockData            testDataForPropertyBootstrap
		errorForCreate      *regexp.Regexp
		errorForUpdate      *regexp.Regexp
		updateChecks        resource.TestCheckFunc
	}{
		"create and remove prefixes - no diff": {
			configPathForCreate: "testdata/TestResPropertyBootstrap/create.tf",
			configPathForUpdate: "testdata/TestResPropertyBootstrap/create_without_prefixes.tf",
			init: func(t *testing.T, papiMock *papi.Mock, iamMock *iam.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(papiMock, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(papiMock, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, data.groupID)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			updateChecks: checkPropertyBootstrapAttributes(testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			}),
		},
		"create and update group id": {
			configPathForCreate: "testdata/TestResPropertyBootstrap/create.tf",
			configPathForUpdate: "testdata/TestResPropertyBootstrap/update_group.tf",
			init: func(t *testing.T, papiMock *papi.Mock, iamMock *iam.Mock, data testDataForPropertyBootstrap) {
				t.Skip("skipping before moving property is enabled again, see DXE-4176")
				ExpectCreateProperty(papiMock, data.name, data.groupID, data.contractID, data.productID,
					data.propertyID).Once()
				prp := &papi.Property{
					AssetID:      "aid_55555",
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(papiMock, data.propertyID, data.groupID, data.contractID, prp).Times(3)

				mockMoveProperty(iamMock, 55555, 1, 111)

				ExpectGetProperty(papiMock, data.propertyID, "grp_111", data.contractID, prp).Twice()
				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, "grp_111")
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			updateChecks: checkPropertyBootstrapAttributes(testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_111",
				contractID: "ctr_2",
				productID:  "prd_3",
			}),
		},
		"create and update name - resource replacement": {
			configPathForCreate: "testdata/TestResPropertyBootstrap/create.tf",
			configPathForUpdate: "testdata/TestResPropertyBootstrap/update_name.tf",
			init: func(t *testing.T, papiMock *papi.Mock, iamMock *iam.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(papiMock, data.name, data.groupID, data.contractID, data.productID,
					data.propertyID).Once()
				prp := &papi.Property{
					AssetID:      "aid_55555",
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(papiMock, data.propertyID, data.groupID, data.contractID, prp).Times(2)

				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, data.groupID).Once()

				ExpectCreateProperty(papiMock, "property_name2", data.groupID, data.contractID, data.productID,
					data.propertyID).Once()

				ExpectGetProperty(papiMock, data.propertyID, data.groupID, data.contractID, prp).Once()

				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, data.groupID).Once()

			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			updateChecks: checkPropertyBootstrapAttributes(testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name2",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			}),
		},
		"create and update name and group id - resource replacement": {
			configPathForCreate: "testdata/TestResPropertyBootstrap/create.tf",
			configPathForUpdate: "testdata/TestResPropertyBootstrap/update_name_and_group.tf",
			init: func(t *testing.T, papiMock *papi.Mock, iamMock *iam.Mock, data testDataForPropertyBootstrap) {
				t.Skip("skipping before moving property is enabled again, see DXE-4176")
				ExpectCreateProperty(papiMock, data.name, data.groupID, data.contractID, data.productID,
					data.propertyID).Once()
				prp := &papi.Property{
					AssetID:      "aid_55555",
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(papiMock, data.propertyID, data.groupID, data.contractID, prp).Times(2)

				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, data.groupID).Once()

				ExpectCreateProperty(papiMock, "property_name2", "grp_93", data.contractID, data.productID,
					data.propertyID).Once()

				ExpectGetProperty(papiMock, data.propertyID, "grp_93", data.contractID, prp).Once()

				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, "grp_93").Once()

			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			updateChecks: checkPropertyBootstrapAttributes(testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name2",
				groupID:    "grp_93",
				contractID: "ctr_2",
				productID:  "prd_3",
			}),
		},
		"create and update contract - error": {
			configPathForCreate: "testdata/TestResPropertyBootstrap/create.tf",
			configPathForUpdate: "testdata/TestResPropertyBootstrap/update_contract.tf",
			init: func(t *testing.T, papiMock *papi.Mock, iamMock *iam.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(papiMock, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(papiMock, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, data.groupID)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			errorForUpdate: regexp.MustCompile("updating field `contract_id` is not possible"),
		},
		"create and update product - error": {
			configPathForCreate: "testdata/TestResPropertyBootstrap/create.tf",
			configPathForUpdate: "testdata/TestResPropertyBootstrap/update_product.tf",
			init: func(t *testing.T, papiMock *papi.Mock, iamMock *iam.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(papiMock, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(papiMock, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, data.groupID)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			errorForUpdate: regexp.MustCompile("updating field `product_id` is not possible"),
		},
		"create and update group - error": {
			// TODO: remove this test after moving property is enabled again, see DXE-4176
			configPathForCreate: "testdata/TestResPropertyBootstrap/create.tf",
			configPathForUpdate: "testdata/TestResPropertyBootstrap/update_group.tf",
			init: func(t *testing.T, papiMock *papi.Mock, iamMock *iam.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(papiMock, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:   "ctr_2",
					GroupID:      "grp_1",
					ProductID:    "prd_3",
					PropertyID:   "prp_123",
					PropertyName: "property_name",
				}
				ExpectGetProperty(papiMock, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(papiMock, data.propertyID, data.contractID, data.groupID)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			errorForUpdate: regexp.MustCompile("updating field `group_id` is not possible"),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			papiMock := &papi.Mock{}
			iamMock := &iam.Mock{}
			if test.init != nil {
				test.init(t, papiMock, iamMock, test.mockData)
			}

			useClient(papiMock, nil, func() {
				useIam(iamMock, func() {
					resource.UnitTest(t, resource.TestCase{
						ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
						IsUnitTest:               true,
						Steps: []resource.TestStep{
							{
								Config:      testutils.LoadFixtureString(t, test.configPathForCreate),
								Check:       checkPropertyBootstrapAttributes(test.mockData),
								ExpectError: test.errorForCreate,
							},
							{
								Config:      testutils.LoadFixtureString(t, test.configPathForUpdate),
								Check:       test.updateChecks,
								ExpectError: test.errorForUpdate,
							},
						},
					})
				})
			})

			papiMock.AssertExpectations(t)
		})
	}
}

func TestBootstrapResourceImport(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		configPath    string
		init          func(*testing.T, *papi.Mock, testDataForPropertyBootstrap)
		mockData      testDataForPropertyBootstrap
		importStateID string
		error         *regexp.Regexp
	}{
		"import with all attributes": {
			configPath: "testdata/TestResPropertyBootstrap/create.tf",
			init: func(t *testing.T, m *papi.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(m, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:    "ctr_2",
					GroupID:       "grp_1",
					ProductID:     "prd_3",
					PropertyID:    "prp_123",
					PropertyName:  "property_name",
					LatestVersion: 1,
				}
				ExpectGetProperty(m, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(m, data.propertyID, data.contractID, data.groupID)
				// import
				ExpectGetProperty(m, data.propertyID, data.groupID, data.contractID, prp)
				ExpectGetPropertyVersion(m, data.propertyID, data.groupID, data.contractID, 1, papi.VersionStatusActive, papi.VersionStatusActive)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			importStateID: "prp_123,2,1",
		},
		"import with only property_id": {
			configPath: "testdata/TestResPropertyBootstrap/create.tf",
			init: func(t *testing.T, m *papi.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(m, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:    "ctr_2",
					GroupID:       "grp_1",
					ProductID:     "prd_3",
					PropertyID:    "prp_123",
					PropertyName:  "property_name",
					LatestVersion: 1,
				}
				ExpectGetProperty(m, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(m, data.propertyID, data.contractID, data.groupID)
				// import
				ExpectGetProperty(m, data.propertyID, "", "", prp)
				ExpectGetPropertyVersion(m, data.propertyID, data.groupID, data.contractID, 1, papi.VersionStatusActive, papi.VersionStatusActive)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			importStateID: "123",
		},
		"import with only property_id and contract_id - error": {
			configPath: "testdata/TestResPropertyBootstrap/create.tf",
			init: func(t *testing.T, m *papi.Mock, data testDataForPropertyBootstrap) {
				ExpectCreateProperty(m, data.name, data.groupID, data.contractID, data.productID, data.propertyID)
				prp := &papi.Property{
					ContractID:    "ctr_2",
					GroupID:       "grp_1",
					ProductID:     "prd_3",
					PropertyID:    "prp_123",
					PropertyName:  "property_name",
					LatestVersion: 1,
				}
				ExpectGetProperty(m, data.propertyID, data.groupID, data.contractID, prp)
				ExpectRemoveProperty(m, data.propertyID, data.contractID, data.groupID)
			},
			mockData: testDataForPropertyBootstrap{
				propertyID: "prp_123",
				name:       "property_name",
				groupID:    "grp_1",
				contractID: "ctr_2",
				productID:  "prd_3",
			},
			importStateID: "123,2",
			error:         regexp.MustCompile("Error: missing group id or contract id"),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := &papi.Mock{}
			if test.init != nil {
				test.init(t, m, test.mockData)
			}

			useClient(m, nil, func() {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
					IsUnitTest:               true,
					Steps: []resource.TestStep{
						{
							Config: testutils.LoadFixtureString(t, "testdata/TestResPropertyBootstrap/create.tf"),
							Check:  checkPropertyBootstrapAttributes(test.mockData),
						},
						{
							ImportState:   true,
							ImportStateId: test.importStateID,
							ResourceName:  "akamai_property_bootstrap.test",
							Check:         checkPropertyBootstrapAttributes(test.mockData),
							ExpectError:   test.error,
						},
					},
				})
			})

			m.AssertExpectations(t)
		})
	}
}

func checkPropertyBootstrapAttributes(data testDataForPropertyBootstrap) resource.TestCheckFunc {
	if data.withoutPrefixes {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "id", data.propertyID),
			resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "group_id", strings.TrimPrefix(data.groupID, "grp_")),
			resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "contract_id", strings.TrimPrefix(data.contractID, "ctr_")),
			resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "product_id", strings.TrimPrefix(data.productID, "prd_")),
			resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "name", data.name))
	}
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "id", str.AddPrefix(data.propertyID, "prp_")),
		resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "group_id", str.AddPrefix(data.groupID, "grp_")),
		resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "contract_id", str.AddPrefix(data.contractID, "ctr_")),
		resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "product_id", str.AddPrefix(data.productID, "prd_")),
		resource.TestCheckResourceAttr("akamai_property_bootstrap.test", "name", data.name),
	)
}
