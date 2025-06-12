describe('Connections Page - Realistic Tests', () => {
  beforeEach(() => {
    // Mock all required API endpoints
    cy.intercept('GET', '/api/connections', {
      statusCode: 200,
      body: [
        {
          id: 'conn-1',
          name: 'Test API Connection',
          integration_id: 'int-1'
        },
        {
          id: 'conn-2',
          name: 'Database Connection',
          integration_id: 'int-2'
        }
      ]
    }).as('getConnections')

    cy.intercept('GET', '/api/integrations', {
      statusCode: 200,
      body: [
        {
          id: 'int-1',
          name: 'Test API',
          credential_type: 'api_key'
        },
        {
          id: 'int-2',
          name: 'PostgreSQL',
          credential_type: 'database'
        }
      ]
    }).as('getIntegrations')

    cy.intercept('GET', '/api/credential-types', {
      statusCode: 200,
      body: [
        {
          type: 'api_key',
          name: 'API Key'
        },
        {
          type: 'database',
          name: 'Database'
        }
      ]
    }).as('getCredentialTypes')

    cy.visit('/connections')
    cy.wait(['@getConnections', '@getIntegrations', '@getCredentialTypes'])
  })

  it('should display the connections page header', () => {
    cy.contains('h2', 'Connections').should('be.visible')
    cy.contains('+ New Connection').should('be.visible')
  })

  it('should show loading state initially', () => {
    // Visit page with slow API response
    cy.intercept('GET', '/api/connections', { delay: 1000, statusCode: 200, body: [] }).as('slowConnections')
    cy.visit('/connections')
    cy.contains('Loading…').should('be.visible')
  })

  it('should display connections in a table', () => {
    cy.get('table').should('be.visible')
    cy.get('thead').within(() => {
      cy.contains('Name').should('be.visible')
      cy.contains('Integration').should('be.visible')
      cy.contains('Actions').should('be.visible')
    })
    
    cy.get('tbody tr').should('have.length', 2)
    cy.contains('Test API Connection').should('be.visible')
    cy.contains('Database Connection').should('be.visible')
    cy.contains('Test API').should('be.visible') // Integration name
    cy.contains('PostgreSQL').should('be.visible')
  })

  it('should open create connection modal', () => {
    cy.contains('+ New Connection').click()
    
    // Modal should appear
    cy.get('.fixed.inset-0').should('be.visible')
    cy.contains('Create Connection').should('be.visible')
    
    // Form fields should be visible
    cy.contains('Integration').should('be.visible')
    cy.contains('Name').should('be.visible')
    cy.get('select').should('be.visible')
    cy.get('input').should('be.visible')
  })

  it('should close modal on cancel', () => {
    cy.contains('+ New Connection').click()
    cy.contains('Cancel').click()
    cy.get('.fixed.inset-0').should('not.exist')
  })

  it('should populate form when integration is selected', () => {
    // Mock credential schema
    cy.intercept('GET', '/api/credential-types/schema/api_key', {
      statusCode: 200,
      body: {
        properties: {
          apiKey: {
            title: 'API Key',
            type: 'string',
            description: 'Your API key'
          },
          baseUrl: {
            title: 'Base URL',
            type: 'string',
            format: 'uri',
            default: 'https://api.example.com'
          }
        },
        required: ['apiKey']
      }
    }).as('getApiKeySchema')

    cy.contains('+ New Connection').click()
    
    // Select integration
    cy.get('select').select('Test API')
    cy.wait('@getApiKeySchema')
    
    // Should show credential fields
    cy.contains('API Key').should('be.visible')
    cy.contains('Base URL').should('be.visible')
    cy.get('input[value="https://api.example.com"]').should('exist') // Default value
    
    // Required field should have asterisk
    cy.contains('API Key *').should('be.visible')
  })

  it('should create a new connection', () => {
    // Mock all necessary APIs
    cy.intercept('GET', '/api/credential-types/schema/api_key', {
      statusCode: 200,
      body: {
        properties: {
          apiKey: { title: 'API Key', type: 'string' }
        },
        required: ['apiKey']
      }
    }).as('getSchema')

    cy.intercept('POST', '/api/credential-types/api_key/test', {
      statusCode: 200,
      body: { success: true }
    }).as('testCredentials')

    cy.intercept('POST', '/api/connections', {
      statusCode: 201,
      body: { id: 'new-conn', name: 'New Connection' }
    }).as('createConnection')

    cy.intercept('GET', '/api/connections', {
      statusCode: 200,
      body: [
        { id: 'conn-1', name: 'Test API Connection', integration_id: 'int-1' },
        { id: 'conn-2', name: 'Database Connection', integration_id: 'int-2' },
        { id: 'new-conn', name: 'New Connection', integration_id: 'int-1' }
      ]
    }).as('getConnectionsUpdated')

    cy.contains('+ New Connection').click()
    cy.get('select').select('Test API')
    cy.wait('@getSchema')
    
    // Use more specific selectors for form fields - name field has no placeholder
    cy.get('input[type="text"]').first().clear().type('My New Connection')
    cy.get('input').filter('[placeholder*="API"], input').last().type('test-api-key')
    
    cy.contains('Test & Save').click()
    cy.wait('@testCredentials')
    cy.wait('@createConnection')
    cy.wait('@getConnectionsUpdated')
    
    // Modal should close and new connection should appear
    cy.get('.fixed.inset-0').should('not.exist')
    cy.contains('New Connection').should('be.visible')
  })

  it('should open edit connection modal', () => {
    // Mock connection details API
    cy.intercept('GET', '/api/connections/conn-1', {
      statusCode: 200,
      body: {
        id: 'conn-1',
        name: 'Test API Connection',
        integration_id: 'int-1',
        secret: {
          baseUrl: 'https://api.example.com',
          apiKey: 'hidden-key'
        }
      }
    }).as('getConnectionDetails')

    cy.intercept('GET', '/api/credential-types/schema/api_key', {
      statusCode: 200,
      body: {
        properties: {
          apiKey: { title: 'API Key', type: 'string' },
          baseUrl: { title: 'Base URL', type: 'string', format: 'uri' }
        },
        required: ['apiKey']
      }
    }).as('getSchema')

    // Click edit button for first connection
    cy.get('tbody tr').first().within(() => {
      cy.contains('Edit').click()
    })

    cy.wait(['@getConnectionDetails', '@getSchema'])
    
    // Should show edit modal
    cy.contains('Edit Connection').should('be.visible')
    cy.get('input[value="Test API Connection"]').should('exist')
    cy.get('input[value="https://api.example.com"]').should('exist') // Non-sensitive field
    
    // Sensitive field should show placeholder
    cy.get('input[placeholder*="current value hidden"]').should('exist')
    
    // Integration should be disabled when editing
    cy.get('select').should('be.disabled')
  })

  it('should delete a connection', () => {
    // Mock window.confirm
    cy.window().then((win) => {
      cy.stub(win, 'confirm').returns(true)
    })

    cy.intercept('DELETE', '/api/connections/conn-1', {
      statusCode: 200,
      body: { success: true }
    }).as('deleteConnection')

    cy.intercept('GET', '/api/connections', {
      statusCode: 200,
      body: [
        { id: 'conn-2', name: 'Database Connection', integration_id: 'int-2' }
      ]
    }).as('getConnectionsAfterDelete')

    // Click delete button for first connection
    cy.get('tbody tr').first().within(() => {
      cy.contains('Delete').click()
    })

    cy.wait('@deleteConnection')
    cy.wait('@getConnectionsAfterDelete')
    
    // Connection should be removed from table
    cy.get('tbody tr').should('have.length', 1)
    cy.contains('Test API Connection').should('not.exist')
  })

  it('should handle credential testing errors', () => {
    cy.intercept('GET', '/api/credential-types/schema/api_key', {
      statusCode: 200,
      body: {
        properties: {
          apiKey: { title: 'API Key', type: 'string' }
        },
        required: ['apiKey']
      }
    }).as('getSchema')

    cy.intercept('POST', '/api/credential-types/api_key/test', {
      statusCode: 401,
      body: { error: 'Invalid API key' }
    }).as('testCredentialsError')

    cy.contains('+ New Connection').click()
    cy.get('select').select('Test API')
    cy.wait('@getSchema')
    
    cy.get('input[type="text"]').first().type('Test Connection')
    cy.get('input').filter('[placeholder*="API"], input').last().type('invalid-key')
    
    cy.contains('Test & Save').click()
    cy.wait('@testCredentialsError')
    
    // Should show error message
    cy.contains('Invalid API key').should('be.visible')
    cy.get('.bg-red-100').should('be.visible')
    
    // Modal should stay open
    cy.get('.fixed.inset-0').should('be.visible')
  })

  it('should handle API errors gracefully', () => {
    // Mock error response
    cy.intercept('GET', '/api/connections', {
      statusCode: 500,
      body: { error: 'Server error' }
    }).as('getConnectionsError')

    cy.visit('/connections')
    cy.wait('@getConnectionsError')
    
    // Should handle error without crashing
    cy.contains('Connections').should('be.visible')
  })
})