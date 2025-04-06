import React, { useState, useEffect } from 'react';
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  CircularProgress,
  Button,
} from '@mui/material';
import {
  Extension as PluginsIcon,
  Folder as FilesIcon,
  Terminal as TerminalIcon,
  Public as ProxyIcon,
} from '@mui/icons-material';
import { Link } from 'react-router-dom';
import Layout from '../components/Layout/Layout';
import apiService from '../services/api';

interface PluginInfo {
  id: string;
  name: string;
  version: string;
  type: string;
  state: string;
}

interface ProxyStatus {
  type: string;
  addr: string;
  running: boolean;
}

const Dashboard: React.FC = () => {
  const [plugins, setPlugins] = useState<PluginInfo[]>([]);
  const [proxyStatus, setProxyStatus] = useState<ProxyStatus[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [pluginsResponse, proxyResponse] = await Promise.all([
          apiService.plugins.list(),
          apiService.proxy.getStatus(),
        ]);
        setPlugins(pluginsResponse.data.data);
        setProxyStatus(proxyResponse.data.data);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return (
      <Layout title="Dashboard">
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh">
          <CircularProgress />
        </Box>
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout title="Dashboard">
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh">
          <Typography color="error">{error}</Typography>
        </Box>
      </Layout>
    );
  }

  const servicePlugins = plugins.filter((plugin) => plugin.type === 'Service');
  const commandPlugins = plugins.filter((plugin) => plugin.type === 'Command');

  return (
    <Layout title="Dashboard">
      <Grid container spacing={3}>
        <Grid item xs={12}>
          <Typography variant="h4" gutterBottom>
            System Overview
          </Typography>
        </Grid>

        {/* Service Plugins */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <PluginsIcon fontSize="large" color="primary" sx={{ mr: 1 }} />
                <Typography variant="h6">Service Plugins</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                {servicePlugins.length} service plugins installed
              </Typography>
              <Box mt={2}>
                <Button component={Link} to="/plugins" variant="outlined" size="small">
                  Manage Plugins
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Command Plugins */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <PluginsIcon fontSize="large" color="secondary" sx={{ mr: 1 }} />
                <Typography variant="h6">Command Plugins</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                {commandPlugins.length} command plugins installed
              </Typography>
              <Box mt={2}>
                <Button component={Link} to="/plugins" variant="outlined" size="small">
                  Manage Plugins
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* File Management */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <FilesIcon fontSize="large" color="primary" sx={{ mr: 1 }} />
                <Typography variant="h6">File Management</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                Upload, download, and manage files
              </Typography>
              <Box mt={2}>
                <Button component={Link} to="/files" variant="outlined" size="small">
                  Manage Files
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Terminal */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <TerminalIcon fontSize="large" color="primary" sx={{ mr: 1 }} />
                <Typography variant="h6">Terminal</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                Create and manage terminal sessions
              </Typography>
              <Box mt={2}>
                <Button component={Link} to="/terminal" variant="outlined" size="small">
                  Open Terminal
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Proxy */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <ProxyIcon fontSize="large" color="primary" sx={{ mr: 1 }} />
                <Typography variant="h6">Proxy Services</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                {proxyStatus.filter((proxy) => proxy.running).length} of {proxyStatus.length} proxy services running
              </Typography>
              <Box mt={2}>
                <Button component={Link} to="/proxy" variant="outlined" size="small">
                  Manage Proxies
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Layout>
  );
};

export default Dashboard;
