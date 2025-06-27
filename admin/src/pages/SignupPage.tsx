import React from 'react';
import { Container, Row, Col, Card, Form, Button } from 'react-bootstrap';

const SignupPage: React.FC = () => (
    <Container className="d-flex vh-100 align-items-center justify-content-center">
        <Row className="w-100 justify-content-center">
            <Col md={4}>
                <Card>
                    <Card.Body>
                        <Card.Title className="mb-4">Sign Up</Card.Title>
                        <Form>
                            <Form.Group className="mb-3" controlId="formBasicEmail">
                                <Form.Label>Email address</Form.Label>
                                <Form.Control type="email" placeholder="Enter email" />
                            </Form.Group>
                            <Form.Group className="mb-3" controlId="formBasicPassword">
                                <Form.Label>Password</Form.Label>
                                <Form.Control type="password" placeholder="Password" />
                            </Form.Group>
                            <Button variant="primary" type="submit" className="w-100">
                                Sign Up
                            </Button>
                        </Form>
                    </Card.Body>
                </Card>
            </Col>
        </Row>
    </Container>
);

export default SignupPage; 