describe('Workflows Page - Realistic Tests', () => {
  beforeEach(() => {
    // Mock the workflows API endpoint
    cy.intercept('GET', '/api/workflows', {
      statusCode: 200,
      body: {
        workflows: [
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
        ],
        total: 2,
        page: 1,
        limit: 20
      }
    }).as('getWorkflows')

    cy.visit('/agents')
    cy.wait('@getWorkflows')
  })

  it('should display the workflows page header', () => {
    cy.contains('h2', 'Workflows').should('be.visible')
    cy.contains('+ New Workflow').should('be.visible')
  })

  it('should show loading state initially', () => {
    // Visit page before API mock is ready
    cy.intercept('GET', '/api/workflows', { 
      delay: 1000, 
      statusCode: 200, 
      body: { workflows: [], total: 0, page: 1, limit: 20 } 
    }).as('slowWorkflows')
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

  it('should open create workflow modal', () => {
    cy.contains('+ New Workflow').click()
    cy.get('.fixed').should('be.visible') // Modal overlay
    cy.contains('Create Workflow').should('be.visible')
    cy.get('input[placeholder=""], input').first().should('be.visible')
    cy.get('textarea').should('be.visible')
  })

  it('should close modal on cancel', () => {
    cy.contains('+ New Workflow').click()
    cy.contains('Cancel').click()
    cy.get('.fixed').should('not.exist')
  })

  it('should create a new workflow', () => {
    cy.intercept('POST', '/api/workflows', {
      statusCode: 201,
      body: { id: 'new-workflow', name: 'New Workflow', description: 'New description' }
    }).as('createWorkflow')

    // Re-mock the GET request to include the new workflow
    cy.intercept('GET', '/api/workflows', {
      statusCode: 200,
      body: {
        workflows: [
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
            id: 'new-workflow',
            name: 'New Workflow',
            description: 'New description'
          }
        ],
        total: 3,
        page: 1,
        limit: 20
      }
    }).as('getWorkflowsWithNew')

    cy.contains('+ New Workflow').click()
    
    // Be more specific about form fields
    cy.get('input[value=""]').first().type('New Workflow')
    cy.get('textarea').type('New description')
    cy.contains('button', 'Create').click()

    cy.wait('@createWorkflow')
    cy.wait('@getWorkflowsWithNew')
    cy.get('.fixed').should('not.exist') // Modal should close
  })

  it('should navigate to workflow builder on row click', () => {
    cy.get('tbody tr').first().click()
    cy.url().should('include', '/workflows/agent-1/edit')
  })

  it('should handle API errors gracefully', () => {
    cy.intercept('GET', '/api/workflows', {
      statusCode: 500,
      body: { error: 'Server error' }
    }).as('getWorkflowsError')

    cy.visit('/agents')
    cy.wait('@getWorkflowsError')
    
    // Should not show loading anymore and should handle the error
    cy.contains('Loading…').should('not.exist')
  })
})