```
Tables:

Shipment:

shipment_id (primary key): Unique identifier for each shipment.
order_id (foreign key): Connects the shipment to a specific order (if applicable).
sender_id (foreign key): References the sender information table.
recipient_id (foreign key): References the recipient information table.
package_details: Description of the package (weight, dimensions, etc.).
shipping_method: Chosen shipping method (air, ground, express, etc.).
shipping_cost: Cost of shipping the package.
status: Current status of the shipment (placed, picked up, in transit, delivered, etc.).
shipped_date: Date the shipment was sent.
estimated_delivery_date: Estimated date for delivery.
tracking_number: Unique tracking number for the shipment (if available).
Sender:

sender_id (primary key): Unique identifier for the sender.
name: Sender's name.
address: Sender's address information.
phone_number: Sender's phone number (optional).
email: Sender's email address (optional).
Recipient:

recipient_id (primary key): Unique identifier for the recipient.
name: Recipient's name.
address: Recipient's address information.
phone_number: Recipient's phone number (optional).
email: Recipient's email address (optional).
```