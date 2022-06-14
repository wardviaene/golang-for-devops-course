package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/wardviaene/golang-for-devops-course/ssh-demo"
)

const location = "westus"

func main() {
	var (
		token  azcore.TokenCredential
		pubKey string
		err    error
	)
	ctx := context.Background()
	subscriptionID := os.Getenv("SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		fmt.Printf("No subscription ID was provided")
		os.Exit(1)
	}
	if pubKey, err = generateKeys(); err != nil {
		fmt.Printf("generatekeys error: %s\n", err)
		os.Exit(1)
	}
	if token, err = getToken(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if err = launchInstance(ctx, subscriptionID, token, &pubKey); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
func generateKeys() (string, error) {
	var (
		privateKey []byte
		publicKey  []byte
		err        error
	)
	if privateKey, publicKey, err = ssh.GenerateKeys(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if err = os.WriteFile("mykey.pem", privateKey, 0600); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if err = os.WriteFile("mykey.pub", publicKey, 0644); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	return string(publicKey), nil
}

func getToken() (azcore.TokenCredential, error) {
	token, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return token, err
	}
	return token, nil
}

func launchInstance(ctx context.Context, subscriptionID string, cred azcore.TokenCredential, pubKey *string) error {
	// create resource client
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}
	resourceGroupParams := armresources.ResourceGroup{
		Location: to.Ptr(location),
	}
	resourcegroupResponse, err := resourceGroupClient.CreateOrUpdate(ctx, "go-azure-demo", resourceGroupParams, nil)
	if err != nil {
		return err
	}

	// create vnet
	virtualNetworkClient, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	vnet, found, err := findVnet(ctx, *resourcegroupResponse.Name, "go-demo", virtualNetworkClient)
	if err != nil {
		return err
	}

	if !found {
		vnetPollerResp, err := virtualNetworkClient.BeginCreateOrUpdate(
			ctx,
			*resourcegroupResponse.Name,
			"go-demo",
			armnetwork.VirtualNetwork{
				Location: to.Ptr(location),
				Properties: &armnetwork.VirtualNetworkPropertiesFormat{
					AddressSpace: &armnetwork.AddressSpace{
						AddressPrefixes: []*string{
							to.Ptr("10.1.0.0/16"),
						},
					},
				},
			},
			nil)

		if err != nil {
			return err
		}

		vnetResponse, err := vnetPollerResp.PollUntilDone(ctx, nil)
		if err != nil {
			return err
		}
		vnet = vnetResponse.VirtualNetwork
	}

	// subnet
	subnetsClient, err := armnetwork.NewSubnetsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	subnetsPollerResp, err := subnetsClient.BeginCreateOrUpdate(
		ctx,
		*resourcegroupResponse.Name,
		*vnet.Name,
		"go-demo",
		armnetwork.Subnet{
			Properties: &armnetwork.SubnetPropertiesFormat{
				AddressPrefix: to.Ptr("10.1.0.0/24"),
			},
		},
		nil,
	)

	if err != nil {
		return err
	}

	subnetResponse, err := subnetsPollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	// Public IP
	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	publicIPPollerResponse, err := publicIPAddressClient.BeginCreateOrUpdate(
		ctx,
		*resourcegroupResponse.Name,
		"go-demo",
		armnetwork.PublicIPAddress{
			Location: to.Ptr(location),
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
			},
		},
		nil,
	)
	if err != nil {
		return err
	}

	publicIPAddressResponse, err := publicIPPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	// Network Security Group
	NewSecurityGroupsClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	networkSecurityGroupPollerResponse, err := NewSecurityGroupsClient.BeginCreateOrUpdate(
		ctx,
		*resourcegroupResponse.Name,
		"go-demo",
		armnetwork.SecurityGroup{
			Location: to.Ptr(location),
			Properties: &armnetwork.SecurityGroupPropertiesFormat{
				SecurityRules: []*armnetwork.SecurityRule{
					{
						Name: to.Ptr("allow-ssh"),
						Properties: &armnetwork.SecurityRulePropertiesFormat{
							SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
							SourcePortRange:          to.Ptr("*"),
							DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
							DestinationPortRange:     to.Ptr("22"),
							Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
							Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
							Description:              to.Ptr("allow ssh on port 22"),
							Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
							Priority:                 to.Ptr(int32(1001)),
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return err
	}
	networkSecurityGroupResponse, err := networkSecurityGroupPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	interfaceClient, err := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	nicPollerResponse, err := interfaceClient.BeginCreateOrUpdate(
		ctx,
		*resourcegroupResponse.Name,
		"go-demo",
		armnetwork.Interface{
			Location: to.Ptr(location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				NetworkSecurityGroup: &armnetwork.SecurityGroup{
					ID: networkSecurityGroupResponse.ID,
				},
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.Ptr("go-demo"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
							Subnet: &armnetwork.Subnet{
								ID: subnetResponse.ID,
							},
							PublicIPAddress: &armnetwork.PublicIPAddress{
								ID: publicIPAddressResponse.ID,
							},
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return err
	}

	nicResponse, err := nicPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	// Create VM
	fmt.Printf("Creating VM:\n")
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(location),
		Identity: &armcompute.VirtualMachineIdentity{
			Type: to.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Offer:     to.Ptr("0001-com-ubuntu-server-focal"),
					Publisher: to.Ptr("canonical"),
					SKU:       to.Ptr("20_04-lts-gen2"),
					Version:   to.Ptr("latest"),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr("go-demo"),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					Caching:      to.Ptr(armcompute.CachingTypesReadWrite),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS), // OSDisk type Standard/Premium HDD/SSD
					},
					DiskSizeGB: to.Ptr[int32](50), // default 127G
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes("Standard_B1s")), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.Ptr("go-demo"),
				AdminUsername: to.Ptr("demo"),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", "demo")),
								KeyData: pubKey,
							},
						},
					},
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: nicResponse.ID,
					},
				},
			},
		},
	}

	pollerResponse, err := vmClient.BeginCreateOrUpdate(ctx, *resourcegroupResponse.Name, "go-demo", parameters, nil)
	if err != nil {
		return err
	}

	vmResponse, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	fmt.Printf("VM Created: %s\n", *vmResponse.ID)

	return nil
}

func findVnet(ctx context.Context, resourceGroupName, vnetName string, vnetClient *armnetwork.VirtualNetworksClient) (armnetwork.VirtualNetwork, bool, error) {
	vnet, err := vnetClient.Get(ctx, resourceGroupName, vnetName, nil)
	if err != nil {
		var errResponse *azcore.ResponseError
		if errors.As(err, &errResponse) && errResponse.ErrorCode == "ResourceNotFound" {
			return vnet.VirtualNetwork, false, nil
		}
		return vnet.VirtualNetwork, false, err
	}

	return vnet.VirtualNetwork, true, nil
}
