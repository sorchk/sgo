import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import {
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Divider,
  Box,
  Typography,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Extension as PluginsIcon,
  Folder as FilesIcon,
  Terminal as TerminalIcon,
  Public as ProxyIcon,
  Code as ShellIcon,
  Logout as LogoutIcon,
} from '@mui/icons-material';
import { useAuth } from '../../contexts/AuthContext';

interface SidebarProps {
  open: boolean;
  onClose: () => void;
  width: number;
}

const Sidebar: React.FC<SidebarProps> = ({ open, onClose, width }) => {
  const location = useLocation();
  const { logout } = useAuth();

  const menuItems = [
    { text: 'Dashboard', icon: <DashboardIcon />, path: '/' },
    { text: 'Plugins', icon: <PluginsIcon />, path: '/plugins' },
    { text: 'Files', icon: <FilesIcon />, path: '/files' },
    { text: 'Terminal', icon: <TerminalIcon />, path: '/terminal' },
    { text: 'Proxy', icon: <ProxyIcon />, path: '/proxy' },
    { text: 'Shell', icon: <ShellIcon />, path: '/shell' },
  ];

  const handleLogout = () => {
    logout();
  };

  return (
    <Drawer
      variant="temporary"
      open={open}
      onClose={onClose}
      sx={{
        width,
        flexShrink: 0,
        '& .MuiDrawer-paper': {
          width,
          boxSizing: 'border-box',
        },
      }}
    >
      <Box sx={{ p: 2 }}>
        <Typography variant="h6" component="div">
          TCP Network Service
        </Typography>
      </Box>
      <Divider />
      <List>
        {menuItems.map((item) => (
          <ListItem
            button
            key={item.text}
            component={Link}
            to={item.path}
            selected={location.pathname === item.path}
            onClick={onClose}
          >
            <ListItemIcon>{item.icon}</ListItemIcon>
            <ListItemText primary={item.text} />
          </ListItem>
        ))}
      </List>
      <Divider />
      <List>
        <ListItem button onClick={handleLogout}>
          <ListItemIcon>
            <LogoutIcon />
          </ListItemIcon>
          <ListItemText primary="Logout" />
        </ListItem>
      </List>
    </Drawer>
  );
};

export default Sidebar;
