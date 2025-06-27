import React from 'react';
import { Container, Row, Col, Navbar, Nav, Dropdown, Image, ListGroup } from 'react-bootstrap';
import DashboardRoutes from './DashboardRoutes';
import { useNavigate } from 'react-router-dom';
import logo from '../assets/sailor-logo.png';

const Dashboard: React.FC = () => {
    const navigate = useNavigate();

    return (
        <Container fluid className="p-0">
            <Navbar style={{ backgroundColor: '#FEF8EE' }} className="px-3">
                <Navbar.Brand style={{ fontWeight: 700, fontSize: 24 }}>
                    <img src={logo} alt="Sailor" style={{ width: 72, height: 72, marginRight: 8 }} />
                    Sailor
                </Navbar.Brand>
                <Nav className="ms-auto">
                    <Dropdown align="end">
                        <Dropdown.Toggle style={{ backgroundColor: '#1C608C' }} id="dropdown-avatar">
                            <Image src="https://ui-avatars.com/api/?name=User" roundedCircle width={32} height={32} />
                        </Dropdown.Toggle>
                        <Dropdown.Menu style={{ backgroundColor: '#FEF8EE' }}>
                            <Dropdown.Item>Logout</Dropdown.Item>
                        </Dropdown.Menu>
                    </Dropdown>
                </Nav>
            </Navbar>
            <Row className="g-0" style={{ minHeight: 'calc(100vh - 56px)', backgroundColor: '#FEF8EE' }}>
                <Col md={2} className="p-0">
                    <ListGroup variant="flush">
                        <ListGroup.Item className="border-0" onClick={() => navigate('/dashboard/apps')} style={{ backgroundColor: '#FEF8EE', }}>
                            <span style={{ marginRight: 8, verticalAlign: 'middle' }}>
                                <svg width="18" height="18" fill="#1C608C" viewBox="0 0 16 16" style={{ marginBottom: 2 }}>
                                    <path d="M4 4h8v8H4z" />
                                    <rect width="16" height="16" fill="none" />
                                </svg>
                            </span>
                            Applications
                        </ListGroup.Item>

                        <ListGroup.Item className="border-0" onClick={() => navigate('/dashboard/settings')} style={{ backgroundColor: '#FEF8EE' }}>
                            <span style={{ marginRight: 8, verticalAlign: 'middle' }}>
                                <svg width="18" height="18" fill="#1C608C" viewBox="0 0 16 16" style={{ marginBottom: 2 }}>
                                    <path d="M8 1a2 2 0 0 1 2 2v1.09a5.978 5.978 0 0 1 2.36.98l.77-.77a2 2 0 1 1 2.83 2.83l-.77.77c.39.73.68 1.53.78 2.36H15a2 2 0 1 1 0 4h-1.09a5.978 5.978 0 0 1-.98 2.36l.77.77a2 2 0 1 1-2.83 2.83l-.77-.77a5.978 5.978 0 0 1-2.36.98V15a2 2 0 1 1-4 0v-1.09a5.978 5.978 0 0 1-2.36-.98l-.77.77a2 2 0 1 1-2.83-2.83l.77-.77A5.978 5.978 0 0 1 1.09 10H1a2 2 0 1 1 0-4h1.09a5.978 5.978 0 0 1 .98-2.36l-.77-.77A2 2 0 1 1 5.13 2.36l.77.77A5.978 5.978 0 0 1 8 3.09V2a2 2 0 0 1 2-2zM8 5a3 3 0 1 0 0 6 3 3 0 0 0 0-6z" />
                                </svg>
                            </span>
                            Settings
                        </ListGroup.Item>
                    </ListGroup>
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