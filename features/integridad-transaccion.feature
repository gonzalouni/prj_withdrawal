# language: en
Feature: Transaction integrity and audit (US-04)
  As a bank (ATM service)
  I want every operation to be atomic and recorded
  So that I guarantee accounting consistency and auditability

  Rule: A transaction completes entirely or it does not happen at all

    Scenario: Successful operation is logged
      Given Alice has 234.56 in their account
      When Alice withdraws 200.00
      Then the transaction is logged with state "EXITOSA"
      And the balance of Alice is 34.56

    Scenario: Declined operation is logged and leaves the balance untouched
      Given Hamza has 198.76 in their account
      When Hamza withdraws 200.00
      Then the transaction is logged with state "RECHAZADA"
      And the balance of Hamza is 198.76

    Scenario: Invalid operation is logged as error
      Given Pedro has 300.00 in their account
      When Pedro withdraws -20.00
      Then the transaction is logged with state "ERROR"
      And the balance of Pedro is 300.00

    Scenario: Failure while persisting rolls everything back
      Given Alice has 234.56 in their account
      And the audit log cannot be written
      When Alice deposits 100.00
      Then the operation fails with state "ERROR"
      And the balance of Alice is 234.56
