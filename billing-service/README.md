```
1. Customers:

customer_id (primary key)
name
company (optional)
email
phone number
address (foreign key to address table - optional if separate)
2. Address (Optional - if storing addresses separately):

address_id (primary key)
line1
line2 (optional)
city
state/province
postal code
country
3. Plans:

plan_id (primary key)
name
description
price
billing_cycle (e.g., monthly, annually)
4. Subscriptions:

subscription_id (primary key)
customer_id (foreign key to Customers table)
plan_id (foreign key to Plans table)
start_date
end_date (optional, if applicable)
status (e.g., active, canceled, paused)
5. Invoices:

invoice_id (primary key)
subscription_id (foreign key to Subscriptions table)
invoice_date
due_date
amount
tax (optional)
status (e.g., issued, paid, overdue)
6. Transactions:

transaction_id (primary key)
invoice_id (foreign key to Invoices table)
payment_method (e.g., credit card, PayPal)
transaction_date
amount
status (e.g., successful, failed)
```