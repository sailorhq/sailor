import React, { useState } from 'react';
import { Container, Row, Col, Card, Form, Button, Alert } from 'react-bootstrap';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const LoginPage: React.FC = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const { login } = useAuth();
    const navigate = useNavigate();

    // Get the page user was trying to access

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            const response = await fetch('http://localhost:7766/api/v1/auth', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'x-username': username,
                    'x-password': password
                },
            });

            if (response.ok) {
                const data = await response.json();
                login(data);
                navigate("/dashboard/apps", { replace: true });
            } else {
                const errorData = await response.json();
                setError(errorData.message || 'Login failed. Please check your credentials.');
            }
        } catch (error) {
            console.error('Login error:', error);
            setError('Network error. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <Container className="d-flex vh-100 align-items-center justify-content-center">
            <Row className="w-400 justify-content-center">
                <Col>
                    <Card className="border-0 shadow-sm" style={{ borderRadius: 10 }}>
                        <Card.Body>
                            <Card.Title className="mb-4">Aye!</Card.Title>
                            {error && (
                                <Alert variant="danger" className="mb-3">
                                    {error}
                                </Alert>
                            )}
                            <Form onSubmit={handleSubmit}>
                                <Form.Group className="mb-3" controlId="formBasicEmail">
                                    <Form.Label>Username</Form.Label>
                                    <Form.Control
                                        placeholder="Enter username"
                                        value={username}
                                        onChange={(e) => setUsername(e.target.value)}
                                        required
                                    />
                                </Form.Group>
                                <Form.Group className="mb-3" controlId="formBasicPassword">
                                    <Form.Label>Password</Form.Label>
                                    <Form.Control
                                        type="password"
                                        placeholder="Password"
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                        required
                                    />
                                </Form.Group>
                                <Button
                                    type="submit"
                                    className="w-100"
                                    style={{ background: '#1C608C', border: 'none' }}
                                    disabled={loading}
                                >
                                    {loading ? 'Logging in...' : 'Login'}
                                </Button>
                            </Form>
                        </Card.Body>
                    </Card>
                </Col>
            </Row>
        </Container>
    );
};

export default LoginPage; 