import React from 'react';
import { Container, Row, Col, Card, Button } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';

const UnauthorizedPage: React.FC = () => {
    const navigate = useNavigate();

    return (
        <Container className="d-flex vh-100 align-items-center justify-content-center">
            <Row className="w-100 justify-content-center">
                <Col md={6}>
                    <Card className="border-0 shadow-sm text-center">
                        <Card.Body>
                            <h1 className="text-danger mb-4">403</h1>
                            <h4 className="mb-3">Access Denied</h4>
                            <p className="text-muted mb-4">
                                You don't have permission to access this page. Please contact your administrator if you believe this is an error.
                            </p>
                            <div className="d-flex gap-2 justify-content-center">
                                <Button
                                    variant="outline-secondary"
                                    onClick={() => navigate(-1)}
                                >
                                    Go Back
                                </Button>
                                <Button
                                    style={{ backgroundColor: '#1C608C', border: 'none' }}
                                    onClick={() => navigate('/dashboard')}
                                >
                                    Go to Dashboard
                                </Button>
                            </div>
                        </Card.Body>
                    </Card>
                </Col>
            </Row>
        </Container>
    );
};

export default UnauthorizedPage; 