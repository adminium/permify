package commands

import (
	"context"
	
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	"github.com/adminium/permify/internal/repositories/mocks"
	"github.com/adminium/permify/internal/schema"
	"github.com/adminium/permify/pkg/database"
	base "github.com/adminium/permify/pkg/pb/base/v1"
	"github.com/adminium/permify/pkg/token"
	"github.com/adminium/permify/pkg/tuple"
)

var _ = Describe("expand-command", func() {
	var expandCommand *ExpandCommand
	
	// DRIVE SAMPLE
	driveSchema := `
	entity user {}
	
	entity organization {
		relation admin @user
	}
	
	entity folder {
		relation org @organization
		relation creator @user
		relation collaborator @user
	
		action read = collaborator
		action update = collaborator
		action delete = creator or org.admin
	}
	
	entity doc {
		relation org @organization
		relation parent @folder
		relation owner @user
	
		action read = (owner or parent.collaborator) or org.admin
		action update = owner and org.admin
		action delete = owner or org.admin
	}
	`
	
	Context("Drive Sample: Expand", func() {
		It("Drive Sample: Case 1", func() {
			var err error
			
			// SCHEMA
			
			schemaReader := new(mocks.SchemaReader)
			
			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, driveSchema)
			Expect(err).ShouldNot(HaveOccurred())
			
			var doc *base.EntityDefinition
			doc, err = schema.GetEntityByName(sch, "doc")
			Expect(err).ShouldNot(HaveOccurred())
			
			var folder *base.EntityDefinition
			folder, err = schema.GetEntityByName(sch, "folder")
			Expect(err).ShouldNot(HaveOccurred())
			
			var organization *base.EntityDefinition
			organization, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())
			
			schemaReader.On("ReadSchemaDefinition", "t1", "doc", "noop").Return(doc, "noop", nil).Times(2)
			schemaReader.On("ReadSchemaDefinition", "t1", "folder", "noop").Return(folder, "noop", nil).Times(1)
			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(1)
			
			// RELATIONSHIPS
			
			relationshipReader := new(mocks.RelationshipReader)
			
			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "owner",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}...), nil).Times(1)
			
			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "parent",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "folder",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)
			
			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "folder",
					Ids:  []string{"1"},
				},
				Relation: "collaborator",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "3",
						Relation: "",
					},
				},
			}...), nil).Times(1)
			
			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "org",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)
			
			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1"},
				},
				Relation: "admin",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}...), nil).Times(1)
			
			expandCommand = NewExpandCommand(schemaReader, relationshipReader)
			
			req := &base.PermissionExpandRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "doc", Id: "1"},
				Permission: "read",
				Metadata: &base.PermissionExpandRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "noop",
				},
			}
			
			var response *base.PermissionExpandResponse
			response, err = expandCommand.Execute(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			
			// fmt.Println(response.GetTree())
			
			Expect(&base.Expand{
				Node: &base.Expand_Expand{
					Expand: &base.ExpandTreeNode{
						Operation: base.ExpandTreeNode_OPERATION_UNION,
						Children: []*base.Expand{
							{
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_UNION,
										Children: []*base.Expand{
											{
												Node: &base.Expand_Leaf{
													Leaf: &base.Result{
														Target: &base.EntityAndRelation{
															Entity: &base.Entity{
																Type: "doc",
																Id:   "1",
															},
															Relation: "owner",
														},
														Subjects: []*base.Subject{
															{
																Type: tuple.USER,
																Id:   "2",
															},
														},
													},
												},
											},
											{
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Node: &base.Expand_Leaf{
																	Leaf: &base.Result{
																		Target: &base.EntityAndRelation{
																			Entity: &base.Entity{
																				Type: "folder",
																				Id:   "1",
																			},
																			Relation: "collaborator",
																		},
																		Subjects: []*base.Subject{
																			{
																				Type: tuple.USER,
																				Id:   "1",
																			},
																			{
																				Type: tuple.USER,
																				Id:   "3",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_UNION,
										Children: []*base.Expand{
											{
												Node: &base.Expand_Leaf{
													Leaf: &base.Result{
														Target: &base.EntityAndRelation{
															Entity: &base.Entity{
																Type: "organization",
																Id:   "1",
															},
															Relation: "admin",
														},
														Subjects: []*base.Subject{
															{
																Type: tuple.USER,
																Id:   "1",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}).Should(Equal(response.Tree))
		})
	})
})
