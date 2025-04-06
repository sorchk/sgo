import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  Chip,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  PlayArrow as StartIcon,
  Stop as StopIcon,
} from '@mui/icons-material';
import Layout from '../components/Layout/Layout';
import apiService from '../services/api';

interface ProxyStatus {
  type: string;
  addr: string;
  running: boolean;
}

const Proxy: React.FC = () => {
  const [proxyStatus, setProxyStatus] = useState<ProxyStatus[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const fetchProxyStatus = async () => {
    try {
      setLoading(true);
      const response = await apiService.proxy.getStatus();
      setProxyStatus(response.data.data);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch proxy status');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProxyStatus();
  }, []);

  const handleStartProxy = async (type: string) => {
    try {
      setActionLoading(type);
      await apiService.proxy.start(type);
      await fetchProxyStatus();
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to start ${type} proxy`);
    } finally {
      setActionLoading(null);
    }
  };

  const handleStopProxy = async (type: string) => {
    try {
      setActionLoading(type);
      await apiService.proxy.stop(type);
      await fetchProxyStatus();
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to stop ${type} proxy`);
    } finally {
      setActionLoading(null);
    }
  };

  if (loading) {
    return (
      <Layout title="Proxy">
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh">
          <CircularProgress />
        </Box>
      </Layout>
    );
  }

  return (
    <Layout title="Proxy">
      <Typography variant="h4" gutterBottom>
        Proxy Services
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Type</TableCell>
              <TableCell>Address</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {proxyStatus.map((proxy) => (
              <TableRow key={proxy.type}>
                <TableCell>{proxy.type}</TableCell>
                <TableCell>{proxy.addr}</TableCell>
                <TableCell>
                  <Chip
                    label={proxy.running ? 'Running' : 'Stopped'}
                    color={proxy.running ? 'success' : 'error'}
                    size="small"
                  />
                </TableCell>
                <TableCell>
                  {proxy.running ? (
                    <Button
                      variant="outlined"
                      color="error"
                      startIcon={<StopIcon />}
                      onClick={() => handleStopProxy(proxy.type)}
                      disabled={actionLoading === proxy.type}
                    >
                      {actionLoading === proxy.type ? <CircularProgress size={24} /> : 'Stop'}
                    </Button>
                  ) : (
                    <Button
                      variant="outlined"
                      color="success"
                      startIcon={<StartIcon />}
                      onClick={() => handleStartProxy(proxy.type)}
                      disabled={actionLoading === proxy.type}
                    >
                      {actionLoading === proxy.type ? <CircularProgress size={24} /> : 'Start'}
                    </Button>
                  )}
                </TableCell>
              </TableRow>
            ))}
            {proxyStatus.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} align="center">
                  No proxy services found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Layout>
  );
};

export default Proxy;
