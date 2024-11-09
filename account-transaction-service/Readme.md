classDiagram
    Account <|-- Transaction
    Account : - ID
    Account : - FirstName
    Account : - LastName
    Account : - CreatedAt
    Account : - Transactions
    Transaction : - ID
    Transaction : - Type
    Transaction : - Amount
    Transaction : - CreatedAt
    Transaction : - Account