/*
Package Clusters enables management and retrieval of Clusters
CCE service.

Example to List Clusters

	listOpts:=clusters.ListOpts{}
	allClusters,err:=clusters.List(client,listOpts)
	if err != nil {
		panic(err)
	}

	for _, cluster := range allClusters {
		fmt.Printf("%+v\n", cluster)
	}

Example to Create a cluster

    createOpts:=clusters.CreateOpts{Kind:"Cluster",
							        ApiVersion:"v3",
							        Metadata:clusters.CreateMetaData{Name:"test-cluster"},
							        Spec:clusters.Spec{Type: "VirtualMachine",
												       Flavor: "cce.s1.small",
												       Version:"v1.7.3-r10",
												       HostNetwork:clusters.HostNetworkSpec{VpcId:"3b9740a0-b44d-48f0-84ee-42eb166e54f7",
																					SubnetId:"3e8e5957-649f-477b-9e5b-f1f75b21c045",},
												       ContainerNetwork:clusters.ContainerNetworkSpec{Mode:"overlay_l2"},
													},
	         }
 	cluster,err := clusters.Create(client,createOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Update a cluster

	updateOpts := clusters.UpdateOpts{Spec:clusters.UpdateSpec{Description:"test"}}

	clusterID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	cluster,err := clusters.Update(client,clusterID,updateOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Delete a cluster

	clusterID := "4e8e5957-649f-477b-9e5b-f1f75b21c03c"

	err := clusters.Delete(client,clusterID).ExtractErr()
	if err != nil {
		panic(err)
	}
*/
package clusters
