import 'dotenv/config'
import fs from 'fs'
import path from 'path'
import { loadConfig } from '../src/config'
import { StartonClient } from '../src/starton'

async function main() {
    const config = loadConfig()
    const wallet = config.signerWallet
    if (!config.startonApiKey || !wallet) throw new Error('STARTON_API_KEY and WALLET_ADDRESS are required')
    const starton = StartonClient.create(config.startonBaseUrl, config.startonApiKey)
    const network = config.defaultNetwork

    const collection = await starton.uploadJson('Starton Playground NFT Collection', {
        name: 'Starton Playground NFT Collection',
        description: 'This NFT collection was created in Starton\'s playground.',
        image: 'https://eu.starton-ipfs.com/ipfs/bafybeicj2womtzppwk7eujerquuhfe5ogtthi4hevmuzsi56cm73agaqim',
    })
    const deployed = await starton.deployFromTemplate({
        templateId: 'ERC721_META_TRANSACTION',
        name: 'My playground smart contract',
        description: 'ERC721 Smart Contract from the Starton Playground',
        network,
        params: ['[Starton Playground] NFT Mint', 'SPNM', 'ipfs://ipfs/', `ipfs://${collection.cid}`, wallet],
        signerWallet: wallet,
    })
    const contractAddress = deployed.smartContract.address
    console.log('Smart contract deployed on address:', contractAddress)

    const image = await starton.uploadFile(fs.readFileSync(path.join(__dirname, 'assets', 'img.png')), 'img.png', 'image/png')
    const metadata = await starton.uploadJson('Starton Playground NFT metadata', {
        name: 'Starton Playground NFT',
        description: 'This NFT was created in Starton\'s playground.',
        image: `https://eu.starton-ipfs.com/ipfs/${image.cid}`,
    })
    const tx = await starton.call({
        network,
        address: contractAddress,
        functionName: 'mint(address,string)',
        params: [wallet, metadata.cid],
        signerWallet: wallet,
    })
    console.log('NFT minted, transaction:', tx.id, tx.transactionHash)
}

main().catch((err) => {
    console.error(err)
    process.exitCode = 1
})
