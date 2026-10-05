// Package identitytest holds the application the kit's tests run for, so they name no real product
// and agree with each other on the one they do name.
package identitytest

import "github.com/oernster/ribbonkit/domain/identity"

// Sample stands for the application the kit runs for in every kit test.
var Sample = identity.App{Name: "SampleRibbon", AppID: "uk.example.SampleRibbon"}
