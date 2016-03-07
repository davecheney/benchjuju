// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package blobstore

var (
	NewResourceCatalog = newResourceCatalog
	NewResource        = newResource
	TxnRunner          = &txnRunner
	PutResourceTxn     = &putResourceTxn
	RequestExpiry      = &requestExpiry
	AfterFunc          = &afterFunc
)

func GetResourceCatalog(ms ManagedStorage) ResourceCatalog {
	return ms.(*managedStorage).resourceCatalog
}

func PutManagedResource(ms ManagedStorage, managedResource ManagedResource, id string) (string, error) {
	return ms.(*managedStorage).putManagedResource(managedResource, id)
}

func ResourceStoragePath(ms ManagedStorage, bucketUUID, user, resourcePath string) (string, error) {
	return ms.(*managedStorage).resourceStoragePath(bucketUUID, user, resourcePath)
}

func RequestQueueLength(ms ManagedStorage) int {
	return len(ms.(*managedStorage).queuedRequests)
}
