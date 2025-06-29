import React, { useEffect, useState } from 'react';
import { Card, ListGroup, Badge } from 'react-bootstrap';
import { getAuditEvents, type AuditEvent } from '../audit';

const AuditPage: React.FC = () => {
    const [auditEvents, setAuditEvents] = useState<AuditEvent[]>([]);

    useEffect(() => {
        getAuditEvents().then(setAuditEvents);
    }, []);

    const formatTimestamp = (timestamp: string) => {
        return new Date(timestamp).toLocaleString();
    };

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                <Card.Title className="mb-4">Audit Trail</Card.Title>
                <ListGroup>
                    {auditEvents.map((event, index) => (
                        <ListGroup.Item key={index} className="d-flex justify-content-between align-items-start">
                            <div className="ms-2 me-auto">
                                <div className="fw-bold">
                                    <Badge bg="primary" className="me-2">{event.action}</Badge>
                                    {event.username}
                                </div>
                                <div className="text-muted small">
                                    {formatTimestamp(event.timestamp)}
                                </div>
                                {event.details && (
                                    <div className="mt-2">
                                        <small className="text-muted">Details:</small>
                                        <pre className="mt-1 mb-0" style={{ fontSize: '0.875rem' }}>
                                            {JSON.stringify(event.details, null, 2)}
                                        </pre>
                                    </div>
                                )}
                            </div>
                        </ListGroup.Item>
                    ))}
                    {auditEvents.length === 0 && (
                        <ListGroup.Item className="text-center text-muted">
                            No audit events found
                        </ListGroup.Item>
                    )}
                </ListGroup>
            </Card.Body>
        </Card>
    );
};

export default AuditPage; 