/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package nri

import (
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/integration"
	"github.com/newrelic/infra-integrations-sdk/v4/log"
)

//This constant is needed only till the workaround to register entity is in place DO NOT MODIFY
const (
	entityTypeInventory = "windowsService"
	heartBeatInventory  = "heartBeat"
)

// ProcessInventory for each entity adds to the inventory entity metadata and metrics dimensions
func ProcessInventory(i *integration.Integration) error {
	entityRules := loadRules()
	for _, e := range i.Entities {
		err := processEntityInventory(e, entityRules)
		if err != nil {
			log.Warn("Error while computing processing entity inventory: " + err.Error())
			return err
		}
	}
	return nil
}

func processEntityInventory(e *integration.Entity, entityRules EntityRules) error {

	err := e.AddInventoryItem(entityTypeInventory, entityRules.EntityName.Label, e.Name())
	if err != nil {
		return err
	}
	for k, v := range e.Metadata.Metadata {
		err = e.AddInventoryItem(entityTypeInventory, k, v)
		if err != nil {
			return err
		}
	}
	for _, m := range e.Metrics {
		for k, v := range m.GetDimensions() {
			err = e.AddInventoryItem(entityTypeInventory, k, v)
			if err != nil {
				return err
			}
		}
	}
	// this is part of the inventory workaround to send entities.
	// Since no metrics from the entities are being send we must update the inventory to keep the entity active.
	err = e.AddInventoryItem(entityTypeInventory, heartBeatInventory, time.Now().Hour())
	if err != nil {
		return err
	}
	return nil
}
