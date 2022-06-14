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

const (
	location = "westeurope"
)

func main() {
	var (
		token  azcore.TokenCredential
		pubKey string
		err    error
	)
	ctx := context.Background()
	subscriptionID := os.Getenv("SUBSCRIPTION_ID")
	if token, err = getToken(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if pubKey, err = generateKeys(); err != nil {
		fmt.Printf("generatekeys error: %s\n", err)
		os.Exit(1)
	}
	if err = launchInstance(ctx, token, subscriptionID, &pubKey); err != nil {
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

func launchInstance(ctx context.Context, token azcore.TokenCredential, subscriptionID string, keydata *string) error {
	var (
		resourceGroupClient *armresources.ResourceGroupsClient
		vnetClient          *armnetwork.VirtualNetworksClient
		err                 error
	)

	if resourceGroupClient, err = armresources.NewResourceGroupsClient(subscriptionID, token, nil); err != nil {
		return err
	}

	resourcegroupParams := armresources.ResourceGroup{
		Location: to.Ptr(location),
	}

	resourceGroup, err := resourceGroupClient.CreateOrUpdate(ctx, "go-demo", resourcegroupParams, nil)
	if err != nil {
		return err
	}

	// vnet
	if vnetClient, err = armnetwork.NewVirtualNetworksClient(subscriptionID, token, nil); err != nil {
		return err
	}
	vnetPparams := armnetwork.VirtualNetwork{
		Location: to.Ptr(location),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					to.Ptr("10.0.0.0/16"), // example 10.1.0.0/16
				},
			},
		},
	}
	vnet, found, err := findVnet(ctx, *resourceGroup.Name, "go-demo", vnetClient)
	if err != nil {
		return err
	}
	if !found {
		vnetPollerResponse, err := vnetClient.BeginCreateOrUpdate(ctx, *resourceGroup.Name, "go-demo", vnetPparams, nil)
		if err != nil {
			return err
		}

		vnetPollerResult, err := vnetPollerResponse.PollUntilDone(ctx, nil)
		if err != nil {
			return err
		}
		vnet = vnetPollerResult.VirtualNetwork
	}
	// create subnets
	subnetClient, err := armnetwork.NewSubnetsClient(subscriptionID, token, nil)
	if err != nil {
		return err
	}

	var subnetResponse armnetwork.SubnetsClientCreateOrUpdateResponse

	subnetParams := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.0.1.0/24"),
		},
	}

	subnetPollerResponse, err := subnetClient.BeginCreateOrUpdate(ctx, *resourceGroup.Name, *vnet.Name, "go-demo", subnetParams, nil)
	if err != nil {
		return err
	}

	subnetResponse, err = subnetPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, token, nil)
	if err != nil {
		return err
	}

	publicIPParams := armnetwork.PublicIPAddress{
		Location: to.Ptr(location),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
		},
	}

	publicIPPollerResponse, err := publicIPAddressClient.BeginCreateOrUpdate(ctx, *resourceGroup.Name, "go-demo", publicIPParams, nil)
	if err != nil {
		return err
	}

	publicIPResponse, err := publicIPPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	// network security group
	nsgClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, token, nil)
	if err != nil {
		return err
	}

	nsgParameters := armnetwork.SecurityGroup{
		Location: to.Ptr(location),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: []*armnetwork.SecurityRule{
				{
					Name: to.Ptr("SSH"), //
					Properties: &armnetwork.SecurityRulePropertiesFormat{
						SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
						SourcePortRange:          to.Ptr("*"),
						DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
						DestinationPortRange:     to.Ptr("22"),
						Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
						Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
						Priority:                 to.Ptr[int32](1001),
						Description:              to.Ptr("inbound port 22 to all"),
						Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
					},
				},
			},
		},
	}

	nsgPollerResponse, err := nsgClient.BeginCreateOrUpdate(ctx, *resourceGroup.Name, "go-demo", nsgParameters, nil)
	if err != nil {
		return err
	}

	nsgResponse, err := nsgPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	// NIC
	nicClient, err := armnetwork.NewInterfacesClient(subscriptionID, token, nil)
	if err != nil {
		return err
	}

	nicParameters := armnetwork.Interface{
		Location: to.Ptr(location),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: to.Ptr("ipConfig"),
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
						Subnet: &armnetwork.Subnet{
							ID: subnetResponse.ID,
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: publicIPResponse.ID,
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: nsgResponse.ID,
			},
		},
	}

	nicPollerResponse, err := nicClient.BeginCreateOrUpdate(ctx, *resourceGroup.Name, "go-demo", nicParameters, nil)
	if err != nil {
		return err
	}

	nicResponse, err := nicPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	// vm client
	fmt.Printf("Launching VM:\n")
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, token, nil)
	if err != nil {
		return err
	}

	vmParams := armcompute.VirtualMachine{
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
						StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS),
					},
					DiskSizeGB: to.Ptr[int32](50),
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes("Standard_B1s")),
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.Ptr("go-demo"),
				AdminUsername: to.Ptr("demo"),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr("/home/demo/.ssh/authorized_keys"),
								KeyData: keydata,
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

	vmPollerResponse, err := vmClient.BeginCreateOrUpdate(ctx, *resourceGroup.Name, "go-demo", vmParams, nil)
	if err != nil {
		return err
	}

	vmResp, err := vmPollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	fmt.Printf("VM created: %s\n", *vmResp.ID)

	return nil
}

func getToken() (azcore.TokenCredential, error) {
	cred, err := azidentity.NewAzureCLICredential(&azidentity.AzureCLICredentialOptions{})
	if err != nil {
		return nil, err
	}

	return cred, nil
}

func findSubnet(ctx context.Context, resourceGroupName, vnetName, subnetName string, subnetClient *armnetwork.SubnetsClient) (bool, error) {
	_, err := subnetClient.Get(ctx, resourceGroupName, vnetName, subnetName, nil)

	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.ErrorCode == "NotFound" {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func findVnet(ctx context.Context, resourceGroupName, vnetName string, vnetClient *armnetwork.VirtualNetworksClient) (armnetwork.VirtualNetwork, bool, error) {
	vnet, err := vnetClient.Get(ctx, resourceGroupName, vnetName, nil)

	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.ErrorCode == "NotFound" {
			return vnet.VirtualNetwork, false, nil
		} else {
			return vnet.VirtualNetwork, false, err
		}
	}
	return vnet.VirtualNetwork, true, nil
}

/*func getTenantID(ctx context.Context, token azcore.TokenCredential) (string, error) {
	var (
		err         error
		accessToken azcore.AccessToken
	)
	if accessToken, err = token.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com"},
	}); err != nil {
		return "", err
	}
	parsedToken, _ := jwt.Parse(accessToken.Token, func(token *jwt.Token) (interface{}, error) {
		// we can't validate this token
		// see https://github.com/AzureAD/azure-activedirectory-identitymodel-extensions-for-dotnet/issues/609#issuecomment-383877585
		return nil, nil
	})
	if tid, ok := parsedToken.Claims.(jwt.MapClaims)["tid"]; ok {
		return tid.(string), nil
	}

	return "", fmt.Errorf("tenant id not found")
}*/
