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
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
} from '@mui/material';
import {
  PlayArrow as StartIcon,
  Stop as StopIcon,
  Info as InfoIcon,
} from '@mui/icons-material';
import Layout from '../components/Layout/Layout';
import apiService from '../services/api';

interface PluginInfo {
  id: string;
  name: string;
  version: string;
  type: string;
  state: string;
}

const Plugins: React.FC = () => {
  const [plugins, setPlugins] = useState<PluginInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [selectedPlugin, setSelectedPlugin] = useState<PluginInfo | null>(null);
  const [pluginDetails, setPluginDetails] = useState<Record<string, string>>({});
  const [detailsOpen, setDetailsOpen] = useState<boolean>(false);

  const fetchPlugins = async () => {
    try {
      setLoading(true);
      const response = await apiService.plugins.list();
      setPlugins(response.data.data);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch plugins');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPlugins();
  }, []);

  const handleStartPlugin = async (id: string) => {
    try {
      setActionLoading(id);
      await apiService.plugins.start(id);
      await fetchPlugins();
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to start plugin ${id}`);
    } finally {
      setActionLoading(null);
    }
  };

  const handleStopPlugin = async (id: string) => {
    try {
      setActionLoading(id);
      await apiService.plugins.stop(id);
      await fetchPlugins();
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to stop plugin ${id}`);
    } finally {
      setActionLoading(null);
    }
  };

  const handleViewDetails = async (plugin: PluginInfo) => {
    try {
      setSelectedPlugin(plugin);
      setActionLoading(plugin.id);
      const response = await apiService.plugins.getInfo(plugin.id);
      setPluginDetails(response.data.data);
      setDetailsOpen(true);
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to get plugin info for ${plugin.id}`);
    } finally {
      setActionLoading(null);
    }
  };

  const handleCloseDetails = () => {
    setDetailsOpen(false);
  };

  if (loading) {
    return (
      <Layout title="Plugins">
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh">
          <CircularProgress />
        </Box>
      </Layout>
    );
  }

  return (
    <Layout title="Plugins">
      <Typography variant="h4" gutterBottom>
        Plugins Management
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
              <TableCell>ID</TableCell>
              <TableCell>Name</TableCell>
              <TableCell>Version</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>State</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {plugins.map((plugin) => (
              <TableRow key={plugin.id}>
                <TableCell>{plugin.id}</TableCell>
                <TableCell>{plugin.name}</TableCell>
                <TableCell>{plugin.version}</TableCell>
                <TableCell>{plugin.type}</TableCell>
                <TableCell>
                  <Chip
                    label={plugin.state}
                    color={plugin.state === 'Enabled' ? 'success' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell>
                  <Box display="flex" gap={1}>
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<InfoIcon />}
                      onClick={() => handleViewDetails(plugin)}
                      disabled={actionLoading === plugin.id}
                    >
                      Info
                    </Button>
                    {plugin.type === 'Service' && (
                      <>
                        {plugin.state === 'Enabled' ? (
                          <Button
                            variant="outlined"
                            size="small"
                            color="error"
                            startIcon={<StopIcon />}
                            onClick={() => handleStopPlugin(plugin.id)}
                            disabled={actionLoading === plugin.id}
                          >
                            {actionLoading === plugin.id ? <CircularProgress size={24} /> : 'Stop'}
                          </Button>
                        ) : (
                          <Button
                            variant="outlined"
                            size="small"
                            color="success"
                            startIcon={<StartIcon />}
                            onClick={() => handleStartPlugin(plugin.id)}
                            disabled={actionLoading === plugin.id}
                          >
                            {actionLoading === plugin.id ? <CircularProgress size={24} /> : 'Start'}
                          </Button>
                        )}
                      </>
                    )}
                  </Box>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={detailsOpen} onClose={handleCloseDetails} maxWidth="md" fullWidth>
        <DialogTitle>Plugin Details: {selectedPlugin?.name}</DialogTitle>
        <DialogContent>
          <DialogContentText component="div">
            <TableContainer>
              <Table>
                <TableBody>
                  {Object.entries(pluginDetails).map(([key, value]) => (
                    <TableRow key={key}>
                      <TableCell component="th" scope="row" sx={{ fontWeight: 'bold' }}>
                        {key}
                      </TableCell>
                      <TableCell>{value}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDetails}>Close</Button>
        </DialogActions>
      </Dialog>
    </Layout>
  );
};

export default Plugins;
