# language: en
Feature: Depositing cash (US-01b)
  As a authenticated customer
  I want to deposit cash indicating an amount
  So that my balance increases and I can see my new balance

  Rule: Customers can only deposit amounts greater than zero

    Scenario: Successful deposit of a positive amount
      Given Alice has 234.56 in their account
      When Alice tries to deposit 100.00
      Then the deposit is successful
      And the balance of Alice is 334.56

    Scenario: Rejected deposit of a negative amount
      Given Hamza has 198.76 in their account
      When Hamza tries to deposit -50.00
      Then the deposit is rejected with a validation error
      And the balance of Hamza is 198.76

    Scenario: Rejected deposit of zero
      Given Carlos has 500.00 in their account
      When Carlos tries to deposit 0.00
      Then the deposit is rejected with a validation error
      And the balance of Carlos is 500.00
