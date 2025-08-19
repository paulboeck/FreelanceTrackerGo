# Freelance Tracker Product Requirements

## Overview
Freelance Tracker will be an application that can be used by any freelancer to keep track of their clients, projects, timesheets, and invoices. It's essentially a CRM with project tracking capabilities.

## Business domain 'objects'
- Settings: The application should have a settings object that the freelancer user can use to update various application-wide settings. The settings object should have the following attributes: default hourly rate.
- Client: the application will be used by a single freelancer to manage details of multiple clients. Each client has a name and address. A client can reside in any country so the application needs to handle US and international addresses.
- Project: each client can have zero to multiple projects. A project should have a name, hourly rate, status. When a new project is created, the hourly rate should be set from a default hourly rate value which is set as part of a global settings object

## Deployment Platform
Initially Freelance Tracker will be a standalone web application that runs on a single workstation. For example, a freelancer user can set up and run the application on their MacBook. Data should be saved to a local/internal SQLLite database. 