# language: en
Feature: Withdrawing cash
  As a authenticated customer
  I want to withdraw cash indicating an amount
  So that I can dispose of my savings and see my new balance

  Rule: Customers cannot withdraw more than their balance

    Scenario: Successful withdrawal within balance
      Given Alice has 234.56 in their account
      When Alice tries to withdraw 200.00
      Then the withdrawal is successful

    Scenario: Declined withdrawal in excess of balance
      Given Hamza has 198.76 in their account
      When Hamza tries to withdraw 200.00
      Then the withdrawal is declined
