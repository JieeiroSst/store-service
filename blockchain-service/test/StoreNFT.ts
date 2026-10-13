import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

describe("StoreNFT", function () {
  async function deployFixture() {
    const [admin, user, other] = await ethers.getSigners();
    const StoreNFT = await ethers.getContractFactory("StoreNFT");
    const nft = await StoreNFT.deploy("Store NFT", "STNFT", "ipfs://", "ipfs://collection", admin.address, admin.address, 500);
    return { nft, admin, user, other };
  }

  it("mints sequential ids with token uri", async function () {
    const { nft, user } = await loadFixture(deployFixture);
    await expect(nft.mint(user.address, "cid-1")).to.emit(nft, "Transfer").withArgs(ethers.constants.AddressZero, user.address, 1);
    await nft.mint(user.address, "cid-2");
    expect(await nft.ownerOf(2)).to.equal(user.address);
    expect(await nft.tokenURI(1)).to.equal("ipfs://cid-1");
    expect(await nft.totalMinted()).to.equal(2);
    expect(await nft.contractURI()).to.equal("ipfs://collection");
  });

  it("restricts minting to minters", async function () {
    const { nft, admin, user } = await loadFixture(deployFixture);
    await expect(nft.connect(user).mint(user.address, "x")).to.be.reverted;
    await nft.grantRole(await nft.MINTER_ROLE(), user.address);
    await nft.connect(user).mint(user.address, "x");
    expect(await nft.balanceOf(user.address)).to.equal(1);
    expect(await nft.hasRole(await nft.MINTER_ROLE(), admin.address)).to.equal(true);
  });

  it("reports royalties and supports interfaces", async function () {
    const { nft, admin, user } = await loadFixture(deployFixture);
    await nft.mint(user.address, "x");
    const [receiver, amount] = await nft.royaltyInfo(1, 10_000);
    expect(receiver).to.equal(admin.address);
    expect(amount).to.equal(500);
    await nft.setTokenRoyalty(1, user.address, 1000);
    expect((await nft.royaltyInfo(1, 10_000))[1]).to.equal(1000);
    expect(await nft.supportsInterface("0x2a55205a")).to.equal(true);
    expect(await nft.supportsInterface("0x80ac58cd")).to.equal(true);
  });

  it("lets only owner or approved burn", async function () {
    const { nft, user, other } = await loadFixture(deployFixture);
    await nft.mint(user.address, "x");
    await expect(nft.connect(other).burn(1)).to.be.revertedWith("StoreNFT: not owner nor approved");
    await nft.connect(user).burn(1);
    await expect(nft.ownerOf(1)).to.be.reverted;
  });

  it("restricts admin setters", async function () {
    const { nft, user } = await loadFixture(deployFixture);
    await expect(nft.connect(user).setBaseURI("x")).to.be.reverted;
    await expect(nft.setBaseURI("https://meta/")).to.emit(nft, "BaseURIUpdated");
    await nft.mint(user.address, "1.json");
    expect(await nft.tokenURI(1)).to.equal("https://meta/1.json");
  });
});
