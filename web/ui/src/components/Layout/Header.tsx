import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  IconButton,
  Box,
} from '@mui/material';
import {
  Menu as MenuIcon,
} from '@mui/icons-material';

interface HeaderProps {
  title: string;
  onMenuClick: () => void;
}

const Header: React.FC<HeaderProps> = ({ title, onMenuClick }) => {
  return (
    <AppBar position="fixed">
      <Toolbar>
        <IconButton
          color="inherit"
          aria-label="open drawer"
          edge="start"
          onClick={onMenuClick}
          sx={{ mr: 2 }}
        >
          <MenuIcon />
        </IconButton>
        <Typography variant="h6" noWrap component="div">
          {title}
        </Typography>
        <Box sx={{ flexGrow: 1 }} />
      </Toolbar>
    </AppBar>
  );
};

export default Header;
