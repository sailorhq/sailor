import React, { useEffect, useState } from 'react';
import { Card, ListGroup, Badge, Button } from 'react-bootstrap';
import { getAuditEvents, type AuditEvent } from '../audit';

const getBadgeColor = (action: string) => {
    switch (action) {
        case 'create_app':
            return 'success';
        default:
            return 'primary';
    }
}

const AuditPage: React.FC = () => {
    const [auditEvents, setAuditEvents] = useState<AuditEvent[]>([]);

    useEffect(() => {
        getAuditEvents().then(setAuditEvents);
    }, []);

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                <Card.Title className="mb-4">Audit Trail</Card.Title>
                <div className="d-flex justify-content-end">
                    <Button className='mb-2' style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={() => getAuditEvents().then(setAuditEvents)}>Refresh</Button>
                </div>
                <ListGroup>
                    {auditEvents.map((event, index) => (
                        <ListGroup.Item key={index} className="d-flex justify-content-between align-items-start">
                            <div className="ms-2 me-auto">
                                <div className="text-muted">
                                    <Badge bg={getBadgeColor(event.action)} className="me-2">{event.action}</Badge>
                                    {JSON.stringify(event, null, 2)}
                                </div>
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