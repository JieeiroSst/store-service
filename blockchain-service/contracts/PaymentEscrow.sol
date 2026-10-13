// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/Address.sol";

contract PaymentEscrow is AccessControl, ReentrancyGuard {
    using SafeERC20 for IERC20;

    bytes32 public constant ARBITER_ROLE = keccak256("ARBITER_ROLE");

    enum Status {
        None,
        Funded,
        Released,
        Refunded
    }

    struct Payment {
        address buyer;
        address seller;
        address token;
        uint256 amount;
        uint64 deadline;
        Status status;
    }

    mapping(bytes32 => Payment) public payments;

    event Funded(bytes32 indexed orderId, address indexed buyer, address indexed seller, address token, uint256 amount, uint64 deadline);
    event Released(bytes32 indexed orderId, address indexed by);
    event Refunded(bytes32 indexed orderId, address indexed by);

    constructor(address admin) {
        require(admin != address(0), "Escrow: admin is zero");
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(ARBITER_ROLE, admin);
    }

    function fund(bytes32 orderId, address seller, address token, uint256 amount, uint64 deadline)
        external
        payable
        nonReentrant
    {
        require(payments[orderId].status == Status.None, "Escrow: order exists");
        require(seller != address(0) && seller != msg.sender, "Escrow: invalid seller");
        require(amount > 0, "Escrow: amount is zero");
        require(deadline > block.timestamp, "Escrow: deadline passed");

        payments[orderId] = Payment(msg.sender, seller, token, amount, deadline, Status.Funded);

        if (token == address(0)) {
            require(msg.value == amount, "Escrow: wrong value");
        } else {
            require(msg.value == 0, "Escrow: unexpected value");
            IERC20 erc20 = IERC20(token);
            uint256 before = erc20.balanceOf(address(this));
            erc20.safeTransferFrom(msg.sender, address(this), amount);
            require(erc20.balanceOf(address(this)) - before == amount, "Escrow: fee-on-transfer token");
        }
        emit Funded(orderId, msg.sender, seller, token, amount, deadline);
    }

    function release(bytes32 orderId) external nonReentrant {
        Payment storage p = payments[orderId];
        require(p.status == Status.Funded, "Escrow: not funded");
        require(msg.sender == p.buyer || hasRole(ARBITER_ROLE, msg.sender), "Escrow: not allowed");
        p.status = Status.Released;
        _pay(p.token, p.seller, p.amount);
        emit Released(orderId, msg.sender);
    }

    function refund(bytes32 orderId) external nonReentrant {
        Payment storage p = payments[orderId];
        require(p.status == Status.Funded, "Escrow: not funded");
        bool allowed = msg.sender == p.seller
            || hasRole(ARBITER_ROLE, msg.sender)
            || (msg.sender == p.buyer && block.timestamp > p.deadline);
        require(allowed, "Escrow: not allowed");
        p.status = Status.Refunded;
        _pay(p.token, p.buyer, p.amount);
        emit Refunded(orderId, msg.sender);
    }

    function _pay(address token, address to, uint256 amount) private {
        if (token == address(0)) {
            Address.sendValue(payable(to), amount);
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }
}
