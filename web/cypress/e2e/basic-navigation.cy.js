describe('Basic Navigation', () => {
  beforeEach(() => {
    cy.mockAPIEndpoints()
  })

  it('should load the home page', () => {
    cy.visit('/')
    cy.contains('Welcome to Agent SaaS').should('be.visible')
    cy.contains('Agent SaaS').should('be.visible')
  })

  it('should have navigation links in header', () => {
    cy.visit('/')
    cy.contains('Agents').should('be.visible')
    cy.contains('Connections').should('be.visible')
  })

  it('should navigate to agents page', () => {
    cy.visit('/')
    cy.contains('Agents').click()
    cy.url().should('include', '/agents')
  })

  it('should navigate to connections page', () => {
    cy.visit('/')
    cy.contains('Connections').click()
    cy.url().should('include', '/connections')
  })

  it('should show active nav state', () => {
    cy.visit('/agents')
    cy.get('nav a[href="/agents"]').should('have.class', 'text-indigo-600')
  })

  it('should return to home page from logo', () => {
    cy.visit('/agents')
    cy.contains('Agent SaaS').click()
    cy.url().should('eq', Cypress.config().baseUrl + '/')
    cy.contains('Welcome to Agent SaaS').should('be.visible')
  })
})