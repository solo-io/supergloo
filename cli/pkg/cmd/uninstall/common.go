package uninstall

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/constants"
)

func staticArgParse(opts *options.Options, installClient *v1.InstallClient) ([]string, error) {
	var meshesToDelete []string
	installList, err := (*installClient).List(constants.SuperglooNamespace, clients.ListOpts{})
	if err != nil {
		return meshesToDelete, fmt.Errorf("unable to retrieve list of installed meshes")
	}
	uop := &opts.Uninstall
	if !uop.All {
		if uop.MeshNames == "" {
			return meshesToDelete, fmt.Errorf("please provide at least one mesh name")
		}

		meshesToDelete = fmtNameList(uop.MeshNames)
		if len(meshesToDelete) < 1 {
			return meshesToDelete, fmt.Errorf("no names supplied, or names incorrectly formatted")
		}
		for _, name := range meshesToDelete {
			_, err := installList.Find(constants.SuperglooNamespace, uop.MeshNames)
			if err != nil {
				return meshesToDelete, fmt.Errorf("supplied mesh name (%s) could not be found", name)
			}
		}

	} else {
		meshesToDelete = activeMeshInstalls(installList)
	}

	return meshesToDelete, nil
}

func dynamicArgParse(opts *options.Options, installClient *v1.InstallClient) ([]string, error) {
	var meshesToDelete []string

	installList, err := (*installClient).List(constants.SuperglooNamespace, clients.ListOpts{})
	if err != nil {
		return meshesToDelete, fmt.Errorf("unable to retrieve list of installed meshes")
	}

	deleteAll, err := deleteAllMeshes()
	if err != nil {
		return meshesToDelete, err
	}

	if !deleteAll {
		meshesToDelete, err = selectMeshByName(activeMeshInstalls(installList))
		if err != nil {
			return meshesToDelete, fmt.Errorf("unable to select list of mesh names")
		}
	} else {
		meshesToDelete = activeMeshInstalls(installList)
	}

	return meshesToDelete, nil
}

func validateArgs(opts *options.Options, installClient *v1.InstallClient) error {
	var meshesToDelete []string
	var err error

	installList, err := (*installClient).List(constants.SuperglooNamespace, clients.ListOpts{})
	if err != nil {
		return fmt.Errorf("unable to retrieve list of install CRDs")
	}

	if len(activeMeshInstalls(installList)) < 1 {
		return fmt.Errorf("no meshes currently installed")
	}

	top := opts.Top

	// if they are using static mode, they must pass all params
	if top.Static {
		meshesToDelete, err = staticArgParse(opts, installClient)
		if err != nil {
			return err
		}
	} else {
		meshesToDelete, err = dynamicArgParse(opts, installClient)
		if err != nil {
			return err
		}
	}

	return uninstallMeshes(meshesToDelete, installClient)
}

func fmtNameList(names string) []string {
	cleanNames := strings.Replace(names, " ", "", -1)
	return strings.Split(cleanNames, ",")
}

func uninstallMeshes(meshList []string, installClient *v1.InstallClient) error {
	installList, err := (*installClient).List(constants.SuperglooNamespace, clients.ListOpts{})
	for i, val := range meshList {
		installCrd, err := installList.Find(constants.SuperglooNamespace, val)
		if err != nil {
			return fmt.Errorf("unable to fetch CRD for (%s) \n finished work: (%s) \n remaining work :", val, meshList[0:i+1], meshList[i:len(meshList)])
		}

		err = updateMeshInstall(installCrd, installClient)
		if err != nil {
			return fmt.Errorf("unable to update CRD for (%s) \n finished work: (%s) \n remaining work :", val, meshList[0:i+1], meshList[i:len(meshList)])
		}

		fmt.Printf("Successfully uninstalled mesh: (%s)", val)

	}
	return err
}

func updateMeshInstall(installCrd *v1.Install, installClient *v1.InstallClient) error {
	installCrd.Enabled = &types.BoolValue{Value: false}
	_, err := (*installClient).Write(installCrd, clients.WriteOpts{OverwriteExisting: true})
	return err
}

func activeMeshInstalls(installList v1.InstallList) []string {
	activeMeshList := make([]string, 0)

	for _, val := range installList {
		if val.Enabled == nil || val.Enabled != nil && val.Enabled.Value {
			activeMeshList = append(activeMeshList, val.Metadata.Name)
		}
	}

	return activeMeshList

}
