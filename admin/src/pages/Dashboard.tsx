import React from 'react';
import { Container, Row, Col, Navbar, Nav, Dropdown, Image, ListGroup, Card } from 'react-bootstrap';
import DashboardRoutes from './DashboardRoutes';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import logo from '../assets/sailor-logo.png';

const Dashboard: React.FC = () => {
    const navigate = useNavigate();
    const { user, logout, hasRole } = useAuth();

    if (!user) {
        logout();
        navigate('/login', { replace: true });
    }

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <Container fluid className="p-0">
            <Navbar style={{ backgroundColor: '#FEF8EE' }} className="px-3">
                <Navbar.Brand style={{ fontWeight: 700, fontSize: 24 }}>
                    <img src={logo} alt="Sailor" style={{ width: 72, height: 72, marginRight: 8 }} />
                    Sailor
                </Navbar.Brand>
                <Nav className="ms-auto">
                    <Dropdown align="end">
                        <Dropdown.Toggle style={{ backgroundColor: '#1C608C', border: 'none' }} id="dropdown-avatar">
                            <Image src={logo} roundedCircle width={32} height={32} />
                            <span style={{ marginLeft: 8 }}>{user?.username}</span>
                        </Dropdown.Toggle>
                        <Dropdown.Menu style={{ backgroundColor: '#FEF8EE' }}>
                            <Dropdown.Item onClick={handleLogout}>Logout</Dropdown.Item>
                        </Dropdown.Menu>
                    </Dropdown>
                </Nav>
            </Navbar>
            <Row className="g-0" style={{ minHeight: 'calc(100vh - 56px)', backgroundColor: '#FEF8EE' }}>
                <Col md={2} className="p-2 mt-3">
                    <Card className="border-0 shadow-sm" style={{ borderRadius: 10 }}>
                        <Card.Body>
                            <ListGroup variant="flush">
                                {hasRole('admin') || hasRole('user') ? <ListGroup.Item className="border-0" onClick={() => navigate('/dashboard/apps')}>
                                    <span style={{ marginRight: 8, verticalAlign: 'middle' }}>
                                        <svg width="18" height="18" fill="#1C608C" viewBox="0 0 16 16" style={{ marginBottom: 2 }}>
                                            <path d="M4 4h8v8H4z" />
                                            <rect width="16" height="16" fill="none" />
                                        </svg>
                                    </span>
                                    Applications
                                </ListGroup.Item> : <></>}

                                {hasRole('admin') ? <ListGroup.Item className="border-0" onClick={() => navigate('/dashboard/settings')}>
                                    <span style={{ marginRight: 8, verticalAlign: 'middle' }}>
                                        <svg width="18" height="18" fill="#1C608C" viewBox="0 0 16 16" style={{ marginBottom: 2 }}>
                                            <path d="M4 4h8v8H4z" />
                                            <rect width="16" height="16" fill="none" />
                                        </svg>
                                    </span>
                                    Settings
                                </ListGroup.Item> : <></>}

                                {hasRole('admin') ? <ListGroup.Item className="border-0" onClick={() => navigate('/dashboard/audit')}>
                                    <span style={{ marginRight: 8, verticalAlign: 'middle' }}>
                                        <svg width="18" height="18" fill="#1C608C" viewBox="0 0 16 16" style={{ marginBottom: 2 }}>
                                            <path d="M4 4h8v8H4z" />
                                            <rect width="16" height="16" fill="none" />
                                        </svg>
                                    </span>
                                    Audit
                                </ListGroup.Item> : <></>}
                            </ListGroup>
                        </Card.Body>
                    </Card>
                </Col>
                <Col md={10} className="p-4">
                    {/* Main content area */}
                    <DashboardRoutes />
                </Col>
            </Row>
        </Container>
    )
};

export default Dashboard; 