// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/interfaces/IERC2981.sol";
import "@openzeppelin/contracts/utils/Address.sol";
import "@openzeppelin/contracts/utils/introspection/ERC165Checker.sol";

contract NFTMarketplace is Ownable, Pausable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    uint96 public constant MAX_FEE_BPS = 1000;

    struct Listing {
        address seller;
        address nft;
        uint256 tokenId;
        address paymentToken;
        uint256 price;
        bool active;
    }

    uint96 public feeBps;
    address public feeRecipient;
    uint256 public nextListingId = 1;
    mapping(uint256 => Listing) public listings;
    mapping(address => mapping(uint256 => uint256)) public activeListingOf;

    event Listed(uint256 indexed listingId, address indexed seller, address indexed nft, uint256 tokenId, address paymentToken, uint256 price);
    event PriceUpdated(uint256 indexed listingId, uint256 price);
    event Cancelled(uint256 indexed listingId);
    event Sold(uint256 indexed listingId, address indexed buyer, uint256 price, uint256 fee, address royaltyReceiver, uint256 royalty);
    event FeeUpdated(uint96 feeBps, address feeRecipient);

    constructor(uint96 feeBps_, address feeRecipient_) {
        _setFee(feeBps_, feeRecipient_);
    }

    function list(address nft, uint256 tokenId, address paymentToken, uint256 price)
        external
        whenNotPaused
        returns (uint256 listingId)
    {
        require(price > 0, "Marketplace: price is zero");
        IERC721 token = IERC721(nft);
        require(token.ownerOf(tokenId) == msg.sender, "Marketplace: not token owner");
        require(
            token.getApproved(tokenId) == address(this) || token.isApprovedForAll(msg.sender, address(this)),
            "Marketplace: not approved"
        );
        uint256 previous = activeListingOf[nft][tokenId];
        if (previous != 0) {
            listings[previous].active = false;
            emit Cancelled(previous);
        }
        listingId = nextListingId++;
        listings[listingId] = Listing(msg.sender, nft, tokenId, paymentToken, price, true);
        activeListingOf[nft][tokenId] = listingId;
        emit Listed(listingId, msg.sender, nft, tokenId, paymentToken, price);
    }

    function updatePrice(uint256 listingId, uint256 price) external {
        Listing storage l = listings[listingId];
        require(l.active, "Marketplace: listing not active");
        require(l.seller == msg.sender, "Marketplace: not seller");
        require(price > 0, "Marketplace: price is zero");
        l.price = price;
        emit PriceUpdated(listingId, price);
    }

    function cancel(uint256 listingId) external {
        Listing storage l = listings[listingId];
        require(l.active, "Marketplace: listing not active");
        require(l.seller == msg.sender || owner() == msg.sender, "Marketplace: not seller");
        _close(listingId, l);
        emit Cancelled(listingId);
    }

    function buy(uint256 listingId, uint256 expectedPrice) external payable nonReentrant whenNotPaused {
        Listing storage l = listings[listingId];
        require(l.active, "Marketplace: listing not active");
        require(l.price == expectedPrice, "Marketplace: price changed");
        require(l.seller != msg.sender, "Marketplace: seller cannot buy");
        IERC721 token = IERC721(l.nft);
        require(token.ownerOf(l.tokenId) == l.seller, "Marketplace: seller no longer owns token");

        address seller = l.seller;
        address paymentToken = l.paymentToken;
        uint256 price = l.price;
        uint256 tokenId = l.tokenId;
        _close(listingId, l);

        uint256 fee = (price * feeBps) / 10_000;
        (address royaltyReceiver, uint256 royalty) = _royalty(address(token), tokenId, price);
        if (royalty + fee > price) {
            royalty = price - fee;
        }
        uint256 sellerAmount = price - fee - royalty;

        if (paymentToken == address(0)) {
            require(msg.value == price, "Marketplace: wrong value");
            _payNative(feeRecipient, fee);
            _payNative(royaltyReceiver, royalty);
            _payNative(seller, sellerAmount);
        } else {
            require(msg.value == 0, "Marketplace: unexpected value");
            IERC20 erc20 = IERC20(paymentToken);
            if (fee > 0) erc20.safeTransferFrom(msg.sender, feeRecipient, fee);
            if (royalty > 0) erc20.safeTransferFrom(msg.sender, royaltyReceiver, royalty);
            erc20.safeTransferFrom(msg.sender, seller, sellerAmount);
        }

        token.safeTransferFrom(seller, msg.sender, tokenId);
        emit Sold(listingId, msg.sender, price, fee, royaltyReceiver, royalty);
    }

    function setFee(uint96 feeBps_, address feeRecipient_) external onlyOwner {
        _setFee(feeBps_, feeRecipient_);
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    function _close(uint256 listingId, Listing storage l) private {
        l.active = false;
        if (activeListingOf[l.nft][l.tokenId] == listingId) {
            delete activeListingOf[l.nft][l.tokenId];
        }
    }

    function _setFee(uint96 feeBps_, address feeRecipient_) private {
        require(feeBps_ <= MAX_FEE_BPS, "Marketplace: fee too high");
        require(feeRecipient_ != address(0) || feeBps_ == 0, "Marketplace: fee recipient is zero");
        feeBps = feeBps_;
        feeRecipient = feeRecipient_;
        emit FeeUpdated(feeBps_, feeRecipient_);
    }

    function _royalty(address nft, uint256 tokenId, uint256 price) private view returns (address, uint256) {
        if (!ERC165Checker.supportsInterface(nft, type(IERC2981).interfaceId)) {
            return (address(0), 0);
        }
        (address receiver, uint256 amount) = IERC2981(nft).royaltyInfo(tokenId, price);
        if (receiver == address(0)) {
            return (address(0), 0);
        }
        return (receiver, amount);
    }

    function _payNative(address to, uint256 amount) private {
        if (amount > 0) {
            Address.sendValue(payable(to), amount);
        }
    }
}
