import { ethers, network } from "hardhat";

async function main() {
  const [deployer] = await ethers.getSigners();
  const admin = process.env.ADMIN_ADDRESS || deployer.address;
  const feeRecipient = process.env.FEE_RECIPIENT || admin;
  const feeBps = Number(process.env.MARKETPLACE_FEE_BPS || 250);
  const royaltyBps = Number(process.env.ROYALTY_BPS || 500);

  const StoreNFT = await ethers.getContractFactory("StoreNFT");
  const nft = await StoreNFT.deploy(
    process.env.NFT_NAME || "Store NFT",
    process.env.NFT_SYMBOL || "STNFT",
    process.env.NFT_BASE_URI || "ipfs://",
    process.env.NFT_CONTRACT_URI || "",
    admin,
    admin,
    royaltyBps
  );
  await nft.deployed();

  const Marketplace = await ethers.getContractFactory("NFTMarketplace");
  const marketplace = await Marketplace.deploy(feeBps, feeRecipient);
  await marketplace.deployed();

  const Escrow = await ethers.getContractFactory("PaymentEscrow");
  const escrow = await Escrow.deploy(admin);
  await escrow.deployed();

  console.log(JSON.stringify({
    network: network.name,
    deployer: deployer.address,
    StoreNFT: nft.address,
    NFTMarketplace: marketplace.address,
    PaymentEscrow: escrow.address,
  }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
