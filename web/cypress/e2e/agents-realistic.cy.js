describe('Agents Page - Realistic Tests', () => {
  beforeEach(() => {
    // Mock the agents API endpoint
    cy.intercept('GET', '/api/agents', {
      statusCode: 200,
      body: [
        {
          id: 'agent-1',
          name: 'Test Workflow',
          description: 'A sample workflow for testing'
        },
        {
          id: 'agent-2', 
          name: 'Data Processing',
          description: 'Processes incoming data'
        }
      ]
    }).as('getAgents')

    cy.visit('/agents')
    cy.wait('@getAgents')
  })

  it('should display the agents page header', () => {
    cy.contains('h2', 'Agents').should('be.visible')
    cy.contains('+ New Agent').should('be.visible')
  })

  it('should show loading state initially', () => {
    // Visit page before API mock is ready
    cy.intercept('GET', '/api/agents', { delay: 1000, statusCode: 200, body: [] }).as('slowAgents')
    cy.visit('/agents')
    cy.contains('Loading…').should('be.visible')
  })

  it('should display agents in a table', () => {
    cy.get('table').should('be.visible')
    cy.get('thead').within(() => {
      cy.contains('Name').should('be.visible')
      cy.contains('Description').should('be.visible')
    })
    
    cy.get('tbody tr').should('have.length', 2)
    cy.contains('Test Workflow').should('be.visible')
    cy.contains('Data Processing').should('be.visible')
  })

  it('should open create agent modal', () => {
    cy.contains('+ New Agent').click()
    cy.get('.fixed').should('be.visible') // Modal overlay
    cy.contains('Create Agent').should('be.visible')
    cy.get('input[placeholder=""], input').first().should('be.visible')
    cy.get('textarea').should('be.visible')
  })

  it('should close modal on cancel', () => {
    cy.contains('+ New Agent').click()
    cy.contains('Cancel').click()
    cy.get('.fixed').should('not.exist')
  })

  it('should create a new agent', () => {
    cy.intercept('POST', '/api/agents', {
      statusCode: 201,
      body: { id: 'new-agent', name: 'New Agent', description: 'New description' }
    }).as('createAgent')

    // Re-mock the GET request to include the new agent
    cy.intercept('GET', '/api/agents', {
      statusCode: 200,
      body: [
        {
          id: 'agent-1',
          name: 'Test Workflow', 
          description: 'A sample workflow for testing'
        },
        {
          id: 'agent-2',
          name: 'Data Processing',
          description: 'Processes incoming data'
        },
        {
          id: 'new-agent',
          name: 'New Agent',
          description: 'New description'
        }
      ]
    }).as('getAgentsWithNew')

    cy.contains('+ New Agent').click()
    
    // Be more specific about form fields
    cy.get('input[value=""]').first().type('New Agent')
    cy.get('textarea').type('New description')
    cy.contains('button', 'Create').click()

    cy.wait('@createAgent')
    cy.wait('@getAgentsWithNew')
    cy.get('.fixed').should('not.exist') // Modal should close
  })

  it('should navigate to agent builder on row click', () => {
    cy.get('tbody tr').first().click()
    cy.url().should('include', '/agents/agent-1/edit')
  })

  it('should handle API errors gracefully', () => {
    cy.intercept('GET', '/api/agents', {
      statusCode: 500,
      body: { error: 'Server error' }
    }).as('getAgentsError')

    cy.visit('/agents')
    cy.wait('@getAgentsError')
    
    // Should not show loading anymore and should handle the error
    cy.contains('Loading…').should('not.exist')
  })
})