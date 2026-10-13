import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

describe("PaymentEscrow", function () {
  const AMOUNT = ethers.utils.parseEther("2");
  const orderId = ethers.utils.id("order-1");

  async function deployFixture() {
    const [admin, buyer, seller, other] = await ethers.getSigners();
    const Escrow = await ethers.getContractFactory("PaymentEscrow");
    const escrow = await Escrow.deploy(admin.address);
    const Token = await ethers.getContractFactory("MockERC20");
    const token = await Token.deploy();
    const deadline = (await time.latest()) + 3600;
    return { escrow, token, admin, buyer, seller, other, deadline };
  }

  it("funds with native coin and releases to seller", async function () {
    const { escrow, buyer, seller, deadline } = await loadFixture(deployFixture);
    await expect(escrow.connect(buyer).fund(orderId, seller.address, ethers.constants.AddressZero, AMOUNT, deadline, { value: AMOUNT }))
      .to.emit(escrow, "Funded")
      .withArgs(orderId, buyer.address, seller.address, ethers.constants.AddressZero, AMOUNT, deadline);
    await expect(escrow.connect(seller).release(orderId)).to.be.revertedWith("Escrow: not allowed");
    await expect(escrow.connect(buyer).release(orderId)).to.changeEtherBalances([escrow, seller], [AMOUNT.mul(-1), AMOUNT]);
    expect((await escrow.payments(orderId)).status).to.equal(2);
    await expect(escrow.connect(buyer).release(orderId)).to.be.revertedWith("Escrow: not funded");
  });

  it("funds with erc20 and arbiter refunds", async function () {
    const { escrow, token, admin, buyer, seller, deadline } = await loadFixture(deployFixture);
    await token.mint(buyer.address, AMOUNT);
    await token.connect(buyer).approve(escrow.address, AMOUNT);
    await escrow.connect(buyer).fund(orderId, seller.address, token.address, AMOUNT, deadline);
    expect(await token.balanceOf(escrow.address)).to.equal(AMOUNT);
    await expect(escrow.connect(admin).refund(orderId)).to.emit(escrow, "Refunded").withArgs(orderId, admin.address);
    expect(await token.balanceOf(buyer.address)).to.equal(AMOUNT);
    expect((await escrow.payments(orderId)).status).to.equal(3);
  });

  it("buyer can refund only after deadline", async function () {
    const { escrow, buyer, seller, deadline } = await loadFixture(deployFixture);
    await escrow.connect(buyer).fund(orderId, seller.address, ethers.constants.AddressZero, AMOUNT, deadline, { value: AMOUNT });
    await expect(escrow.connect(buyer).refund(orderId)).to.be.revertedWith("Escrow: not allowed");
    await time.increaseTo(deadline + 1);
    await expect(escrow.connect(buyer).refund(orderId)).to.changeEtherBalance(buyer, AMOUNT);
  });

  it("seller can refund voluntarily, others cannot", async function () {
    const { escrow, buyer, seller, other, deadline } = await loadFixture(deployFixture);
    await escrow.connect(buyer).fund(orderId, seller.address, ethers.constants.AddressZero, AMOUNT, deadline, { value: AMOUNT });
    await expect(escrow.connect(other).refund(orderId)).to.be.revertedWith("Escrow: not allowed");
    await expect(escrow.connect(seller).refund(orderId)).to.changeEtherBalance(buyer, AMOUNT);
  });

  it("validates funding", async function () {
    const { escrow, buyer, seller, deadline } = await loadFixture(deployFixture);
    const zero = ethers.constants.AddressZero;
    await expect(escrow.connect(buyer).fund(orderId, buyer.address, zero, AMOUNT, deadline, { value: AMOUNT })).to.be.revertedWith("Escrow: invalid seller");
    await expect(escrow.connect(buyer).fund(orderId, seller.address, zero, 0, deadline)).to.be.revertedWith("Escrow: amount is zero");
    await expect(escrow.connect(buyer).fund(orderId, seller.address, zero, AMOUNT, 1, { value: AMOUNT })).to.be.revertedWith("Escrow: deadline passed");
    await expect(escrow.connect(buyer).fund(orderId, seller.address, zero, AMOUNT, deadline, { value: 1 })).to.be.revertedWith("Escrow: wrong value");
    await escrow.connect(buyer).fund(orderId, seller.address, zero, AMOUNT, deadline, { value: AMOUNT });
    await expect(escrow.connect(buyer).fund(orderId, seller.address, zero, AMOUNT, deadline, { value: AMOUNT })).to.be.revertedWith("Escrow: order exists");
  });
});
