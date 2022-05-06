/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/ava-labs/avalanche-cli/pkg/models"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var readCmd = &cobra.Command{
	Use:   "describe",
	Short: "Print a summary of the subnet’s configuration",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run:  readGenesis,
	Args: cobra.ExactArgs(1),
}

var printGenesisOnly *bool

func init() {
	subnetCmd.AddCommand(readCmd)

	printGenesisOnly = readCmd.Flags().BoolP(
		"genesis",
		"g",
		false,
		"Print the genesis to the console directly instead of the summary",
	)
}

func printGenesis(subnetName string) error {
	usr, _ := user.Current()
	genesisFile := filepath.Join(usr.HomeDir, BaseDir, subnetName+genesis_suffix)
	gen, err := os.ReadFile(genesisFile)
	if err != nil {
		return err
	}
	fmt.Println(string(gen))
	return nil
}

func printGasTable(genesis core.Genesis) {

	const art = `
  _____              _____             __ _
 / ____|            / ____|           / _(_)
| |  __  __ _ ___  | |     ___  _ __ | |_ _  __ _
| | |_ |/ _` + `  / __| | |    / _ \| '_ \|  _| |/ _` + `  |
| |__| | (_| \__ \ | |___| (_) | | | | | | | (_| |
 \_____|\__,_|___/  \_____\___/|_| |_|_| |_|\__, |
                                             __/ |
                                            |___/
`

	fmt.Printf(art)
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"Gas Parameter", "Value"}
	table.SetHeader(header)
	table.SetRowLine(true)

	table.Append([]string{"GasLimit", genesis.Config.FeeConfig.GasLimit.String()})
	table.Append([]string{"MinBaseFee", genesis.Config.FeeConfig.MinBaseFee.String()})
	table.Append([]string{"TargetGas", genesis.Config.FeeConfig.TargetGas.String()})
	table.Append([]string{"BaseFeeChangeDenominator", genesis.Config.FeeConfig.BaseFeeChangeDenominator.String()})
	table.Append([]string{"MinBlockGasCost", genesis.Config.FeeConfig.MinBlockGasCost.String()})
	table.Append([]string{"MaxBlockGasCost", genesis.Config.FeeConfig.MaxBlockGasCost.String()})
	table.Append([]string{"TargetBlockRate", strconv.FormatUint(genesis.Config.FeeConfig.TargetBlockRate, 10)})
	table.Append([]string{"BlockGasCostStep", genesis.Config.FeeConfig.BlockGasCostStep.String()})

	table.Render()
}

func printAirdropTable(genesis core.Genesis) {
	const art = `
          _         _
    /\   (_)       | |
   /  \   _ _ __ __| |_ __ ___  _ __
  / /\ \ | | '__/ _` + `  | '__/ _ \| '_ \
 / ____ \| | | | (_| | | | (_) | |_) |
/_/    \_\_|_|  \__,_|_|  \___/| .__/
                               | |
                               |_|
`
	fmt.Printf(art)
	if len(genesis.Alloc) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		header := []string{"Address", "Airdrop Amount (10^18)", "Airdrop Amount (wei)"}
		table.SetHeader(header)
		table.SetRowLine(true)

		for address := range genesis.Alloc {
			amount := genesis.Alloc[address].Balance
			formattedAmount := new(big.Int).Div(amount, big.NewInt(params.Ether))
			table.Append([]string{address.Hex(), formattedAmount.String(), amount.String()})
		}

		table.Render()
	} else {
		fmt.Printf("No airdrops allocated")
	}
}

func printPrecompileTable(genesis core.Genesis) {
	const art = `

  _____                                    _ _
 |  __ \                                  (_) |
 | |__) | __ ___  ___ ___  _ __ ___  _ __  _| | ___  ___
 |  ___/ '__/ _ \/ __/ _ \| '_ ` + `  _ \| '_ \| | |/ _ \/ __|
 | |   | | |  __/ (_| (_) | | | | | | |_) | | |  __/\__ \
 |_|   |_|  \___|\___\___/|_| |_| |_| .__/|_|_|\___||___/
                                    | |
                                    |_|

`
	fmt.Printf(art)

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"Precompile", "Admin"}
	table.SetHeader(header)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.SetRowLine(true)

	precompileSet := false

	// Native Minting
	// timestamp := genesis.Config.ContractNativeMinterConfig.BlockTimestamp
	for _, address := range genesis.Config.ContractNativeMinterConfig.AllowListAdmins {
		table.Append([]string{"Native Minter", address.Hex()})
		precompileSet = true
	}

	// Contract allow list
	for _, address := range genesis.Config.ContractDeployerAllowListConfig.AllowListAdmins {
		table.Append([]string{"Contract Allow list", address.Hex()})
		precompileSet = true
	}

	if precompileSet {
		table.Render()
	} else {
		fmt.Println("No precompiles set")
	}
}

func describeSubnetEvmGenesis(subnetName string, sc models.Sidecar) error {
	// Load genesis
	genesis, err := loadEvmGenesis(subnetName)
	if err != nil {
		return err
	}

	// Write gas table
	printGasTable(genesis)
	// fmt.Printf("\n\n")
	printAirdropTable(genesis)
	printPrecompileTable(genesis)
	return nil
}

func readGenesis(cmd *cobra.Command, args []string) {
	subnetName := args[0]
	if *printGenesisOnly {
		err := printGenesis(subnetName)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		// read in sidecar
		sc, err := loadSidecar(subnetName)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch sc.Vm {
		case models.SubnetEvm:
			err = describeSubnetEvmGenesis(subnetName, sc)
		default:
			fmt.Println("Unknown genesis format for", sc.Vm)
			fmt.Println("Printing genesis")
			err = printGenesis(subnetName)
		}
		if err != nil {
			fmt.Println(err)
		}
	}
}