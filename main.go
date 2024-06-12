package main

import (
	"blobstreamx-example/client"
	shareloader "blobstreamx-example/contracts/ShareLoader.sol"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tendermint/tendermint/rpc/client/http"
)

func main() {
	// step 0. get eth-rpc and trpc endpoint
	ethEndpoint := "https://arbitrum-sepolia-rpc.publicnode.com"
	trpcEndpoint := "https://celestia-mocha-rpc.publicnode.com:443"
	// trpcEndpoint := "https://rpc.celestia-mocha.com:443"
	// trpcEndpoint := "https://26658-calindra-celestianode-p9zxr391sw1.ws-us114.gitpod.io:443"

	// See contract here: https://sepolia.arbiscan.io/contract/0xf2787995D9eb43b57eAdB361227Ddf4FEC99b5Df
	// This is the address of the ShareLoader contract
	// The contract wraps the DAVerifier.verifySharesToDataRootTupleRoot method
	shareloaderAddress := common.HexToAddress("0xf2787995D9eb43b57eAdB361227Ddf4FEC99b5Df")

	// step 1: connect to eth and trpc endpoints
	eth, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		panic(fmt.Errorf("failed to connect to the Ethereum node: %w", err))
	}
	trpc, err := http.New(trpcEndpoint, "/websocket")
	if err != nil {
		panic(fmt.Errorf("failed to connect to the Tendermint RPC: %w", err))
	}
	fmt.Println("Successfully connected to the Ethereum node and Tendermint RPC")

	// step 2: generate share proof
	sp := &client.SharePointer{
		Height: 2034386, //
		Start:  10,
		End:    11,
	}

	// sp := &client.SharePointer{
	// 	Height: 2034689, //
	// 	Start:  22,
	// 	End:    23,
	// }

	// sp := &client.SharePointer{
	// 	Height: 2027691,
	// 	Start:  0,
	// 	End:    1,
	// }
	// sp := &client.SharePointer{
	// 	Height: 1490181,
	// 	Start:  1,
	// 	End:    14,
	// }

	proof, root, err := client.GetShareProof(eth, trpc, sp)
	if err != nil {
		panic(fmt.Errorf("failed to get share proof: %w", err))
	}

	fmt.Printf("\n\nINI - The info that we need to make a GioRequest\n")
	fmt.Printf("* Namespace %s\n", common.Bytes2Hex(proof.Namespace.Id[:]))
	fmt.Printf("* Height %d\n", proof.AttestationProof.Tuple.Height.Uint64())
	fmt.Printf("* Start %d\n", proof.ShareProofs[0].BeginKey.Uint64())
	fmt.Printf("END - The info that we need to make a GioRequest\n\n")

	fmt.Printf("Data len %d\n %s\n", len(proof.Data), common.Bytes2Hex(proof.Data[0]))

	fmt.Printf("Successfully generated share proof %s\n len = %d\n; Digest = %s;\n Min = %s;\nMax = %s;\nshare proof len = %d\n\n",
		common.Bytes2Hex(proof.AttestationProof.Tuple.DataRoot[:]),
		len(proof.RowRoots),
		common.Bytes2Hex(proof.RowRoots[0].Digest[:]),
		common.Bytes2Hex(proof.RowRoots[0].Min.Id[:]),
		common.Bytes2Hex(proof.RowRoots[0].Max.Id[:]),
		len(proof.ShareProofs),
		// proof.ShareProofs[0].SideNodes,
	)

	// step 3: verify the share proof
	loader, err := shareloader.NewShareloader(shareloaderAddress, eth)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate shareloader contract: %w", err))
	}

	valid, errCodes, err := loader.VerifyShares(nil, *proof, root)
	if err != nil {
		panic(fmt.Errorf("failed to verify share proof: %w", err))
	}

	// step 4: print the result
	if !valid {
		fmt.Println("Proof is invalid", "Error codes:", errCodes)
		os.Exit(1)
	}
	fmt.Println("Proof is valid")
}
