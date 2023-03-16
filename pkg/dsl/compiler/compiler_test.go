package compiler

import (
	"errors"
	"testing"
	
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	"github.com/adminium/permify/pkg/dsl/parser"
	base "github.com/adminium/permify/pkg/pb/base/v1"
)

// TestCompiler -
func TestCompiler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "compiler-suite")
}

var _ = Describe("compiler", func() {
	Context("NewCompiler", func() {
		It("Case 1", func() {
			sch, err := parser.NewParser(`
			entity user {}`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(true, sch)
			
			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())
			
			Expect(is).Should(Equal([]*base.EntityDefinition{
				{
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
			}))
		})
		
		It("Case 2", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = owner or admin
			}
			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())
			
			i := []*base.EntityDefinition{
				{
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Actions: map[string]*base.ActionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "admin",
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
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}
			
			Expect(is).Should(Equal(i))
		})
		
		It("Case 3", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = owner or (admin and owner)
			}
			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())
			
			i := []*base.EntityDefinition{
				{
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Actions: map[string]*base.ActionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Rewrite{
													Rewrite: &base.Rewrite{
														RewriteOperation: base.Rewrite_OPERATION_INTERSECTION,
														Children: []*base.Child{
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Exclusion: false,
																		Type: &base.Leaf_ComputedUserSet{
																			ComputedUserSet: &base.ComputedUserSet{
																				Relation: "admin",
																			},
																		},
																	},
																},
															},
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Exclusion: false,
																		Type: &base.Leaf_ComputedUserSet{
																			ComputedUserSet: &base.ComputedUserSet{
																				Relation: "owner",
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
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}
			
			Expect(is).Should(Equal(i))
		})
		
		It("Case 4", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = owner
			}
			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())
			
			i := []*base.EntityDefinition{
				{
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Actions: map[string]*base.ActionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Exclusion: false,
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "owner",
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}
			
			Expect(is).Should(Equal(i))
		})
		
		It("Case 5", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = maintainer or admin
			}
			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			_, err = c.Compile()
			Expect(err).Should(Equal(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())))
		})
		
		It("Case 6", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity parent {
				
				relation admin @user
			}

			entity organization {
				
				relation parent @parent
				relation admin @user
			}

			entity repository {
				
				relation parent @organization
				action update = parent.parent.admin or admin
			}
			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			_, err = c.Compile()
			Expect(err).Should(Equal(errors.New(base.ErrorCode_ERROR_CODE_NOT_SUPPORTED_RELATION_WALK.String())))
		})
		
		It("Case 7", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = owner or admin
			}

			entity repository {
				
				relation parent @organization
				relation owner @user

				action delete = owner or (parent.update or not parent.owner)
			}

			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())
			
			i := []*base.EntityDefinition{
				{
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Actions: map[string]*base.ActionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "admin",
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
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
				{
					Name: "repository",
					Actions: map[string]*base.ActionDefinition{
						"delete": {
							Name: "delete",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Rewrite{
													Rewrite: &base.Rewrite{
														RewriteOperation: base.Rewrite_OPERATION_UNION,
														Children: []*base.Child{
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Exclusion: false,
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "update",
																				},
																			},
																		},
																	},
																},
															},
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Exclusion: true,
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "owner",
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
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"parent": {
							Name: "parent",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"delete": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}
			
			Expect(is).Should(Equal(i))
		})
		
		It("Case 8", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = owner or admin
			}

			entity repository {
				
				relation parent @organization
				relation owner @user @organization#admin @organization#owner

				action delete = owner or (parent.update or not parent.owner)
			}

			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())
			
			i := []*base.EntityDefinition{
				{
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Actions: map[string]*base.ActionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "admin",
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
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
				{
					Name: "repository",
					Actions: map[string]*base.ActionDefinition{
						"delete": {
							Name: "delete",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Rewrite{
													Rewrite: &base.Rewrite{
														RewriteOperation: base.Rewrite_OPERATION_UNION,
														Children: []*base.Child{
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Exclusion: false,
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "update",
																				},
																			},
																		},
																	},
																},
															},
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Exclusion: true,
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "owner",
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
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"parent": {
							Name: "parent",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
								{
									Type:     "organization",
									Relation: "admin",
								},
								{
									Type:     "organization",
									Relation: "owner",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"delete": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}
			
			Expect(is).Should(Equal(i))
		})
		
		It("Case 9", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = owner or admin
			}

			entity repository {
				
				relation parent @organization
				relation owner @user @organization

				action delete = owner or (parent.update or not parent.owner)
			}

			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			_, err = c.Compile()
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_MUST_HAVE_ONE_ENTITY_REFERENCE.String()))
		})
		
		It("Case 10", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = owner or admin
			}

			entity repository {
				
				relation parent @organization
				relation owner @user @organization#update

				action delete = owner or (parent.update or not parent.owner)
			}

			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			_, err = c.Compile()
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String()))
		})
		
		It("Case 11", func() {
			sch, err := parser.NewParser(`
			entity user {
    			relation org @organization

    			action read = org.admin
    			action write = org.admin
			}

			entity organization {
    			relation admin @user
			}

			entity division {
    			relation manager @user @organization#admin

				action read = manager
    			action write = manager
			}
			`).Parse()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			c := NewCompiler(false, sch)
			
			var is []*base.EntityDefinition
			is, err = c.Compile()
			
			Expect(err).ShouldNot(HaveOccurred())
			
			i := []*base.EntityDefinition{
				{
					Name: "user",
					Relations: map[string]*base.RelationDefinition{
						"org": {
							Name: "org",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
					},
					Actions: map[string]*base.ActionDefinition{
						"read": {
							Name: "read",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Exclusion: false,
										Type: &base.Leaf_TupleToUserSet{
											TupleToUserSet: &base.TupleToUserSet{
												TupleSet: &base.TupleSet{
													Relation: "org",
												},
												Computed: &base.ComputedUserSet{
													Relation: "admin",
												},
											},
										},
									},
								},
							},
						},
						"write": {
							Name: "write",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Exclusion: false,
										Type: &base.Leaf_TupleToUserSet{
											TupleToUserSet: &base.TupleToUserSet{
												TupleSet: &base.TupleSet{
													Relation: "org",
												},
												Computed: &base.ComputedUserSet{
													Relation: "admin",
												},
											},
										},
									},
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"org":   base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"read":  base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
						"write": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
				{
					Name: "organization",
					Relations: map[string]*base.RelationDefinition{
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					Actions: map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{
						"admin": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					},
				},
				{
					Name: "division",
					Relations: map[string]*base.RelationDefinition{
						"manager": {
							Name: "manager",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
								{
									Type:     "organization",
									Relation: "admin",
								},
							},
						},
					},
					Actions: map[string]*base.ActionDefinition{
						"read": {
							Name: "read",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Exclusion: false,
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "manager",
											},
										},
									},
								},
							},
						},
						"write": {
							Name: "write",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Exclusion: false,
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "manager",
											},
										},
									},
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"manager": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"read":    base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
						"write":   base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}
			
			Expect(is).Should(Equal(i))
		})
	})
})
