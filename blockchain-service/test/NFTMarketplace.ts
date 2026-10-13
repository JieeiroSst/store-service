import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

describe("NFTMarketplace", function () {
  const PRICE = ethers.utils.parseEther("1");

  async function deployFixture() {
    const [owner, creator, seller, buyer, treasury] = await ethers.getSigners();
    const StoreNFT = await ethers.getContractFactory("StoreNFT");
    const nft = await StoreNFT.deploy("Store NFT", "STNFT", "ipfs://", "", owner.address, creator.address, 500);
    const Marketplace = await ethers.getContractFactory("NFTMarketplace");
    const market = await Marketplace.deploy(250, treasury.address);
    const Token = await ethers.getContractFactory("MockERC20");
    const token = await Token.deploy();
    await nft.mint(seller.address, "cid");
    await nft.connect(seller).setApprovalForAll(market.address, true);
    return { nft, market, token, owner, creator, seller, buyer, treasury };
  }

  it("sells for native coin and splits fee and royalty", async function () {
    const { nft, market, seller, buyer, creator, treasury } = await loadFixture(deployFixture);
    await expect(market.connect(seller).list(nft.address, 1, ethers.constants.AddressZero, PRICE)).to.emit(market, "Listed");
    const fee = PRICE.mul(250).div(10_000);
    const royalty = PRICE.mul(500).div(10_000);
    await expect(market.connect(buyer).buy(1, PRICE, { value: PRICE }))
      .to.emit(market, "Sold")
      .withArgs(1, buyer.address, PRICE, fee, creator.address, royalty)
      .and.to.changeEtherBalances([buyer, treasury, creator, seller], [PRICE.mul(-1), fee, royalty, PRICE.sub(fee).sub(royalty)]);
    expect(await nft.ownerOf(1)).to.equal(buyer.address);
    expect((await market.listings(1)).active).to.equal(false);
  });

  it("sells for erc20", async function () {
    const { nft, market, token, seller, buyer, creator, treasury } = await loadFixture(deployFixture);
    await market.connect(seller).list(nft.address, 1, token.address, PRICE);
    await token.mint(buyer.address, PRICE);
    await token.connect(buyer).approve(market.address, PRICE);
    await expect(market.connect(buyer).buy(1, PRICE, { value: 1 })).to.be.revertedWith("Marketplace: unexpected value");
    await market.connect(buyer).buy(1, PRICE);
    expect(await token.balanceOf(treasury.address)).to.equal(PRICE.mul(250).div(10_000));
    expect(await token.balanceOf(creator.address)).to.equal(PRICE.mul(500).div(10_000));
    expect(await token.balanceOf(seller.address)).to.equal(PRICE.mul(9250).div(10_000));
    expect(await nft.ownerOf(1)).to.equal(buyer.address);
  });

  it("validates listing", async function () {
    const { nft, market, seller, buyer } = await loadFixture(deployFixture);
    await expect(market.connect(buyer).list(nft.address, 1, ethers.constants.AddressZero, PRICE)).to.be.revertedWith("Marketplace: not token owner");
    await expect(market.connect(seller).list(nft.address, 1, ethers.constants.AddressZero, 0)).to.be.revertedWith("Marketplace: price is zero");
    await nft.connect(seller).setApprovalForAll(market.address, false);
    await expect(market.connect(seller).list(nft.address, 1, ethers.constants.AddressZero, PRICE)).to.be.revertedWith("Marketplace: not approved");
  });

  it("relisting cancels the previous listing", async function () {
    const { nft, market, seller, buyer } = await loadFixture(deployFixture);
    await market.connect(seller).list(nft.address, 1, ethers.constants.AddressZero, PRICE);
    await expect(market.connect(seller).list(nft.address, 1, ethers.constants.AddressZero, PRICE.mul(2))).to.emit(market, "Cancelled").withArgs(1);
    await expect(market.connect(buyer).buy(1, PRICE, { value: PRICE })).to.be.revertedWith("Marketplace: listing not active");
    expect(await market.activeListingOf(nft.address, 1)).to.equal(2);
  });

  it("protects buyer from price change and stale listings", async function () {
    const { nft, market, seller, buyer, owner } = await loadFixture(deployFixture);
    await market.connect(seller).list(nft.address, 1, ethers.constants.AddressZero, PRICE);
    await market.connect(seller).updatePrice(1, PRICE.mul(2));
    await expect(market.connect(buyer).buy(1, PRICE, { value: PRICE })).to.be.revertedWith("Marketplace: price changed");
    await expect(market.connect(buyer).buy(1, PRICE.mul(2), { value: PRICE })).to.be.revertedWith("Marketplace: wrong value");
    await nft.connect(seller).transferFrom(seller.address, owner.address, 1);
    await expect(market.connect(buyer).buy(1, PRICE.mul(2), { value: PRICE.mul(2) })).to.be.revertedWith("Marketplace: seller no longer owns token");
  });

  it("cancel by seller or owner only", async function () {
    const { nft, market, seller, buyer, owner } = await loadFixture(deployFixture);
    await market.connect(seller).list(nft.address, 1, ethers.constants.AddressZero, PRICE);
    await expect(market.connect(buyer).cancel(1)).to.be.revertedWith("Marketplace: not seller");
    await expect(market.connect(owner).cancel(1)).to.emit(market, "Cancelled").withArgs(1);
    await expect(market.connect(seller).cancel(1)).to.be.revertedWith("Marketplace: listing not active");
  });

  it("owner controls fee and pause", async function () {
    const { nft, market, seller, buyer } = await loadFixture(deployFixture);
    await expect(market.connect(buyer).setFee(100, buyer.address)).to.be.revertedWith("Ownable: caller is not the owner");
    await expect(market.setFee(1001, buyer.address)).to.be.revertedWith("Marketplace: fee too high");
    await market.pause();
    await expect(market.connect(seller).list(nft.address, 1, ethers.constants.AddressZero, PRICE)).to.be.revertedWith("Pausable: paused");
  });
});
