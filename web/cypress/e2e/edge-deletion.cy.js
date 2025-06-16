describe('Edge Deletion Functionality', () => {
  beforeEach(() => {
    // Visit a workflow builder page with some nodes and edges
    cy.visit('http://localhost:5173/agents/test-agent/edit');
  });

  it('should load the builder page', () => {
    // Wait for the page to load and verify basic elements
    cy.get('[data-testid="react-flow"]', { timeout: 15000 }).should('be.visible');
    cy.get('.react-flow__renderer').should('be.visible');
  });

  it('should have edge hover functionality', () => {
    // Wait for the page to load
    cy.get('[data-testid="react-flow"]', { timeout: 15000 }).should('be.visible');
    
    // Check if there are any edges and test hover functionality
    cy.get('.react-flow__edge').then($edges => {
      if ($edges.length > 0) {
        // Hover over the first edge
        cy.get('.react-flow__edge').first().trigger('mouseover');
        
        // Look for delete button (may be in edge label renderer)
        cy.get('button[title="Delete edge"]', { timeout: 3000 }).should('exist');
      } else {
        cy.log('No edges found to test hover functionality');
      }
    });
  });
});